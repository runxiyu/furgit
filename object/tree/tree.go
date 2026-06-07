package tree

import (
	"slices"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/tree/mode"
)

// Tree represents a fully materialized Git tree object.
//
// The zero value is an empty tree.
// Entries are stored sorted by Git tree ordering and are duplicate-free;
// use [Tree.Insert] to add entries and [Tree.Entries] to read them.
//
// Labels: MT-Unsafe.
type Tree struct {
	entries []Entry
}

// Entry represents a single entry in a tree.
type Entry struct {
	Mode mode.Mode
	Name string
	ID   id.ObjectID
}

// Entries returns a copy of the tree's entries in Git tree order.
//
// Mutating the returned slice does not affect the tree.
//
// Labels: Life-Independent.
func (tree *Tree) Entries() []Entry {
	return slices.Clone(tree.entries)
}
