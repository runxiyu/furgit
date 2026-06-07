// Package tree provides parsed tree objects and tree serialization.
//
// A [Tree] is a fully materialized Git tree object.
// It keeps its entries sorted and duplicate-free internally;
// callers build trees with [Tree.Insert] and read them with [Tree.Entries],
// neither of which exposes the internal storage.
package tree
