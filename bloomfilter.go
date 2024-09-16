// Package bloomfilter is face-meltingly fast, thread-safe,
// marshalable, unionable, probability- and
// optimal-size-calculating Bloom filter in go
//
// https://github.com/steakknife/bloomfilter
//
// # Copyright © 2014, 2015, 2018 Barry Allard
//
// MIT license
package bloomfilter

import (
	"hash"
	"sync"
	"sync/atomic"
)

// Filter is an opaque Bloom filter type
type Filter struct {
	// The RLock semantics are different from usual:
	// "RLock" is for actions (including writes) that do not require a consistent
	// view of the bits
	// "Lock" is for actions (including reads) that do require a consistent view
	// of the bits
	lock sync.RWMutex
	bits []atomic.Uint64 // mutable
	keys []uint64        // immutable after init
	m    uint64          // number of bits the "bits" field should recognize; immutable after init
	n    atomic.Uint64   // number of inserted elements; mutable
}

func (f *Filter) getBits() []uint64 {
	out := make([]uint64, len(f.bits))
	for i, v := range f.bits {
		out[i] = v.Load()
	}
	return out
}

func (f *Filter) setBits(b []uint64) {
	for i, v := range b[0:min(len(b), len(f.bits))] {
		f.bits[i].Store(v)
	}
}

// Hashable -> hashes
func (f *Filter) hash(v hash.Hash64) []uint64 {
	rawHash := v.Sum64()
	n := len(f.keys)
	hashes := make([]uint64, n)
	for i := 0; i < n; i++ {
		hashes[i] = rawHash ^ f.keys[i]
	}
	return hashes
}

// M is the size of Bloom filter, in bits
func (f *Filter) M() uint64 {
	return f.m
}

// K is the count of keys
func (f *Filter) K() uint64 {
	return uint64(len(f.keys))
}

// Add a hashable item, v, to the filter
func (f *Filter) Add(v hash.Hash64) {
	h := f.hash(v)
	f.lock.RLock()
	defer f.lock.RUnlock()
	for _, i := range h {
		// f.setBit(i)
		i %= f.m
		f.bits[i>>6].Or(1 << uint(i&0x3f))
	}
	f.n.Add(1)
}

// AddC adds a hashable item, v, to the filter, testing for its presence
// beforehand.
// false: f definitely does not contain value v
// true:  f maybe contains value v
func (f *Filter) AddC(v hash.Hash64) bool {
	h := f.hash(v)
	f.lock.RLock()
	f.lock.RUnlock()
	r := uint64(1)
	for _, i := range h {
		i %= f.m
		r &= (f.bits[i>>6].Load() >> uint(i&0x3f)) & 1
		f.bits[i>>6].Or(1 << uint(i&0x3f))
	}
	f.n.Add(1)
	return uint64ToBool(r)
}

// Contains tests if f contains v
// false: f definitely does not contain value v
// true:  f maybe contains value v
func (f *Filter) Contains(v hash.Hash64) bool {
	h := f.hash(v)
	f.lock.RLock()
	defer f.lock.RUnlock()
	r := uint64(1)
	for _, i := range h {
		// r |= f.getBit(k)
		i %= f.m
		r &= (f.bits[i>>6].Load() >> uint(i&0x3f)) & uint64(1)
	}
	return uint64ToBool(r)
}

// Copy f to a new Bloom filter
func (f *Filter) Copy() (*Filter, error) {
	out, err := f.NewCompatible()
	if err != nil {
		return nil, err
	}
	f.lock.Lock()
	defer f.lock.Unlock()
	out.setBits(f.getBits())
	out.n.Store(f.n.Load())
	return out, nil
}

// UnionInPlace merges Bloom filter f2 into f
func (f *Filter) UnionInPlace(f2 *Filter) error {
	f.lock.RLock()
	defer f.lock.RUnlock()
	f2.lock.RLock()
	defer f2.lock.RUnlock()

	if err := f.verifyCompatible(f2); err != nil {
		return err
	}

	for i, word := range f2.bits {
		f.bits[i].Or(word.Load())
	}
	return nil
}

// Union merges f2 and f2 into a new Filter out
func (f *Filter) Union(f2 *Filter) (out *Filter, err error) {
	f.lock.RLock()
	defer f.lock.RUnlock()
	f2.lock.RLock()
	defer f2.lock.RUnlock()
	if err := f.verifyCompatible(f2); err != nil {
		return nil, err
	}
	out, err = f.NewCompatible()
	if err != nil {
		return nil, err
	}
	for i, word := range f2.bits {
		out.bits[i].Store(f.bits[i].Load() | word.Load())
	}
	return out, nil
}
