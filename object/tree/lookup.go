package tree

import (
	"bytes"
	"slices"

	"lindenii.org/go/furgit/object/tree/mode"
)

// Find returns the entry with the given name, if present.
//
// A name matches whether stored as a blob-like or as a subtree entry,
// so both orderings are searched.
//
// The returned entry is a shallow copy:
// its Name aliases the tree's internal storage,
// so it must not be mutated and shares the tree's lifetime.
//
// Labels: Life-Parent, Mut-No.
func (tree *Tree) Find(name []byte) (Entry, bool) {
	for _, searchIsTree := range [...]bool{true, false} {
		index, ok := slices.BinarySearchFunc(tree.entries, name, func(existing Entry, target []byte) int {
			return nameCompare(existing.Name, existing.Mode == mode.Directory, target, searchIsTree)
		})
		if ok && bytes.Equal(tree.entries[index].Name, name) {
			return tree.entries[index], true
		}
	}

	return Entry{}, false //exhaustruct:ignore
}
