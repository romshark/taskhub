// Package slices provides common functions for slice-operations.
package slices

import "golang.org/x/exp/slices"

// WithoutIndex returns s without with the item at index i removed.
// Examples:
//
//	s=[1,2,3]; i=0 // [2,3]
//	s=[1,2,3]; i=1 // [1,3]
//	s=[1]; i=0 // []
func WithoutIndex[T any](s []T, i int) []T {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

// FilterInPlace returns a new slice containing only those elements
// for which predicate returned true.
// Examples:
//
//	s=[1,2,3,4]; predicate = func(x int) bool { return x%2==0 } // [2,4]
func FilterInPlace[T any](s []T, predicate func(T) bool) []T {
	i := 0
	for x := range s {
		if predicate(s[x]) {
			s[i] = s[x]
			i++
		}
	}
	return s[:i]
}

// IsSubset returns true if a is a subset of b,
// in other words: b contains all values of a;
// Examples:
//
//	a=[1,2]; b=[1,2,3] // true
//	a=[2,1,1]; b=[1,2,3] // true
//	a=[1]; b=[1,2,3] // true
//	a=[]; b=[1,2,3] // true
//	a=[1,5]; b=[1,2,3] // false
func IsSubset[T comparable](a, b []T) bool {
ON_A:
	for ia := range a {
		for ib := range b {
			if b[ib] == a[ia] {
				continue ON_A
			}
		}
		return false
	}
	return true
}

// Contains returns true if s contains x, otherwise returns false.
func Contains[T comparable](s []T, x T) bool {
	for i := range s {
		if s[i] == x {
			return true
		}
	}
	return false
}

// IsSubsetGet returns true if a is a subset of b,
// in other words: b contains all values of a;
// Examples:
//
//	a=[1,2]; b=[1,2,3] // true
//	a=[2,1,1]; b=[1,2,3] // true
//	a=[1]; b=[1,2,3] // true
//	a=[]; b=[1,2,3] // true
//	a=[1,5]; b=[1,2,3] // false
func IsSubsetGet[T1 comparable, T2 any](
	a []T1, b []T2, get func(T2) T1,
) bool {
ON_A:
	for ia := range a {
		for ib := range b {
			if get(b[ib]) == a[ia] {
				continue ON_A
			}
		}
		return false
	}
	return true
}

// Copy returns a shallow copy of s.
func Copy[T any](s []T) []T {
	if s == nil {
		return nil
	}
	c := make([]T, len(s))
	copy(c, s)
	return c
}

// SortAndLimit applies sortFnLess if != nil and limit if != nil to s.
func SortAndLimit[T any](s []T, sortFnLess func(a, b T) bool, limit int) []T {
	if s == nil {
		return nil
	}
	if sortFnLess != nil {
		slices.SortFunc(s, sortFnLess)
	}
	if limit >= 0 && limit < len(s) {
		s = s[:limit]
	}
	return s
}

// AppendUnique returns s with x appended if s didn't contain x.
func AppendUnique[T comparable](s []T, x T) []T {
	for i := range s {
		if s[i] == x {
			return s
		}
	}
	return append(s, x)
}
