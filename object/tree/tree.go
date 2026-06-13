package tree

import (
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
	Name []byte
	ID   id.ObjectID
}

// Entries returns the tree's entries in Git tree order.
//
// The returned slice aliases the tree's internal storage,
// so it must not be mutated,
// and it is invalidated by any subsequent call that mutates the tree,
// such as [Tree.Insert].
// Use [Tree.Clone] for an independent tree.
//
// Labels: Life-Parent, Mut-No.
func (tree *Tree) Entries() []Entry {
	return tree.entries
}
