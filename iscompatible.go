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
	"fmt"
)

func uint64ToBool(x uint64) bool {
	return x != 0
}

// returns 0 if equal, does not compare len(b0) with len(b1)
func noBranchCompareUint64s(b0, b1 []uint64) uint64 {
	r := uint64(0)
	for i, b0i := range b0 {
		r |= b0i ^ b1[i]
	}
	return r
}

// IsCompatible is true if f and f2 can be Union()ed together
func (f *Filter) IsCompatible(f2 *Filter) bool {
	f.lock.RLock()
	defer f.lock.RUnlock()

	f2.lock.RLock()
	defer f2.lock.RUnlock()
	return f.isCompatible(f2)
}

// IsCompatible is true if f and f2 can be Union()ed together
func (f *Filter) isCompatible(f2 *Filter) bool {
	return f.M() == f2.M() &&
		f.K() == f2.K() &&
		noBranchCompareUint64s(f.keys, f2.keys) == 0
}

func (f *Filter) verifyCompatible(f2 *Filter) error {
	f.lock.RLock()
	defer f.lock.RUnlock()

	f2.lock.RLock()
	defer f2.lock.RUnlock()
	if f.isCompatible(f2) {
		return nil
	}
	e := make([]string, 3)
	if f.M() != f2.M() {
		e[0] = fmt.Sprintf("M=%d and M=%d", f.K(), f2.K())
	}
	if f.K() != f2.K() {
		e[1] = fmt.Sprintf("K=%d and K=%d", f.K(), f2.K())
	}
	if noBranchCompareUint64s(f.keys, f2.keys) != 0 {
		e[2] = "Mismatched Keys"
	}
	return errIncompatibleBloomFilters(e)
}
