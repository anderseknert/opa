// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	"fmt"
	"maps"
	"slices"

	"github.com/open-policy-agent/opa/v1/util"
)

// VarSet represents a set of variables.
type VarSet map[Var]struct{}

// NewVarSet returns a new VarSet containing the specified variables.
func NewVarSet(vs ...Var) VarSet {
	// overallocating slightly cheaper than growing
	s := make(VarSet, len(vs)*2)
	for _, v := range vs {
		s[v] = struct{}{}
	}
	return s
}

// Add updates the set to include the variable "v".
func (s VarSet) Add(v Var) {
	s[v] = struct{}{}
}

// Contains returns true if the set contains the variable "v".
func (s VarSet) Contains(v Var) bool {
	if s == nil {
		return false
	}
	_, ok := s[v]
	return ok
}

// Copy returns a shallow copy of the VarSet.
func (s VarSet) Copy() VarSet {
	cpy := VarSet{}
	for v := range s {
		cpy.Add(v)
	}
	return cpy
}

// Diff returns a VarSet containing variables in s that are not in vs.
func (s VarSet) Diff(vs VarSet) VarSet {
	if s == nil {
		return nil
	}
	if vs == nil {
		return s.Copy()
	}

	r := VarSet{}
	for v := range s {
		_, ok := vs[v]
		if !ok {
			r[v] = struct{}{}
		}
	}
	return r
}

// Equal returns true if s contains exactly the same elements as vs.
func (s VarSet) Equal(vs VarSet) bool {
	if len(s.Diff(vs)) > 0 {
		return false
	}
	return len(vs.Diff(s)) == 0
}

// Intersect returns a VarSet containing variables in s that are in vs.
func (s VarSet) Intersect(vs VarSet) VarSet {
	if s == nil {
		return nil
	}

	r := VarSet{}
	for v := range s {
		_, ok := vs[v]
		if ok {
			r[v] = struct{}{}
		}
	}
	return r
}

// Sorted returns a sorted slice of vars from s. If s is nil, Sorted returns nil.
func (s VarSet) Sorted() []Var {
	if s == nil {
		return nil
	}
	sorted := make([]Var, 0, len(s))
	for v := range s {
		sorted = append(sorted, v)
	}
	slices.SortFunc(sorted, VarCompare)
	return sorted
}

// Update merges the other VarSet into this VarSet.
func (s VarSet) Update(vs VarSet) {
	maps.Copy(s, vs)
}

// OrEmpty returns the VarSet if it is not nil. Otherwise, it returns an empty VarSet.
func (s VarSet) OrEmpty() VarSet {
	if s == nil {
		return VarSet{}
	}
	return s
}

// Clear removes all variables from the set.
func (s VarSet) Clear() {
	clear(s)
}

func (s VarSet) String() string {
	return fmt.Sprintf("%v", util.KeysSorted(s))
}
