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
	"strings"
)

func errHash() error {
	return fmt.Errorf(
		"Hash mismatch, the Bloom filter is probably corrupt")
}
func errK() error {
	return fmt.Errorf(
		"keys must have length %d or greater", KMin)
}
func errM() error {
	return fmt.Errorf(
		"m (number of bits in the Bloom filter) must be >= %d", MMin)
}
func errUniqueKeys() error {
	return fmt.Errorf(
		"Bloom filter keys must be unique")
}

type errIncompatible struct {
	s []string
}

var ErrIncompatible *errIncompatible = &errIncompatible{s: []string{"example"}}

func (e *errIncompatible) Error() string {
	out := make([]string, 0, 3)
	for _, i := range e.s {
		if i != "" {
			out = append(out, i)
		}
	}
	return fmt.Sprintf("Cannot perform union on two incompatible Bloom filters: %s", strings.Join(out, ", "))
}

func (e *errIncompatible) Is(err error) bool {
	_, ok := err.(*errIncompatible)
	return ok
}

func errIncompatibleBloomFilters(s []string) error {
	return &errIncompatible{s: s}
}
