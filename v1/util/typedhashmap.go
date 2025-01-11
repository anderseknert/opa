// Copyright 2025 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"strings"
)

type hasher interface {
	Hash() int
}

type typedHashEntry[K hasher, V any] struct {
	k    K
	v    V
	next *typedHashEntry[K, V]
}

// TypedTypedHashMap represents a key/value map.
type TypedHashMap[K hasher, V any] struct {
	eq    func(K, K) bool
	table map[int]*typedHashEntry[K, V]
	size  int
}

// NewTypedHashMap returns a new empty TypedHashMap.
func NewTypedHashMap[K hasher, V any](eq func(K, K) bool) *TypedHashMap[K, V] {
	return &TypedHashMap[K, V]{
		eq:   eq,
		size: 0,
	}
}

// Copy returns a shallow copy of this TypedHashMap.
func (h *TypedHashMap[K, V]) Copy() *TypedHashMap[K, V] {
	cpy := NewTypedHashMap[K, V](h.eq)
	h.Iter(func(k K, v V) bool {
		cpy.Put(k, v)
		return false
	})
	return cpy
}

// Get returns the value for k.
func (h *TypedHashMap[K, V]) Get(k K) (V, bool) {
	if h.table == nil {
		var zero V

		return zero, false
	}

	hash := k.Hash()
	for entry := h.table[hash]; entry != nil; entry = entry.next {
		if h.eq(entry.k, k) {
			return entry.v, true
		}
	}

	var zero V

	return zero, false
}

// Delete removes the key k.
func (h *TypedHashMap[K, V]) Delete(k K) {
	if h.table == nil {
		return
	}

	hash := k.Hash()
	var prev *typedHashEntry[K, V]
	for entry := h.table[hash]; entry != nil; entry = entry.next {
		if h.eq(entry.k, k) {
			if prev != nil {
				prev.next = entry.next
			} else {
				h.table[hash] = entry.next
			}
			h.size--
			return
		}
		prev = entry
	}
}

// Iter invokes the iter function for each element in the TypedHashMap.
// If the iter function returns true, iteration stops and the return value is true.
// If the iter function never returns true, iteration proceeds through all elements
// and the return value is false.
func (h *TypedHashMap[K, V]) Iter(iter func(k K, v V) bool) bool {
	if h == nil || h.table == nil {
		return false
	}

	for _, entry := range h.table {
		for ; entry != nil; entry = entry.next {
			if iter(entry.k, entry.v) {
				return true
			}
		}
	}

	return false
}

// Len returns the current size of this TypedHashMap.
func (h *TypedHashMap[K, V]) Len() int {
	return h.size
}

// Put inserts a key/value pair into this TypedHashMap. If the key is already present, the existing
// value is overwritten.
func (h *TypedHashMap[K, V]) Put(k K, v V) {
	if h.table == nil {
		h.table = map[int]*typedHashEntry[K, V]{}
	}

	hash := k.Hash()
	head := h.table[hash]
	for entry := head; entry != nil; entry = entry.next {
		if h.eq(entry.k, k) {
			entry.v = v
			return
		}
	}
	h.table[hash] = &typedHashEntry[K, V]{k: k, v: v, next: head}
	h.size++
}

func (h *TypedHashMap[K, V]) String() string {
	var buf []string
	h.Iter(func(k K, v V) bool {
		buf = append(buf, fmt.Sprintf("%v: %v", k, v))
		return false
	})
	return "{" + strings.Join(buf, ", ") + "}"
}

// Update returns a new TypedHashMap with elements from the other TypedHashMap put into this TypedHashMap.
// If the other TypedHashMap contains elements with the same key as this TypedHashMap, the value
// from the other TypedHashMap overwrites the value from this TypedHashMap.
func (h *TypedHashMap[K, V]) Update(other *TypedHashMap[K, V]) *TypedHashMap[K, V] {
	updated := h.Copy()
	other.Iter(func(k K, v V) bool {
		updated.Put(k, v)
		return false
	})
	return updated
}
