package tree

import (
	"slices"

	"lindenii.org/go/furgit/object/tree/mode"
)

// Find returns the entry with the given name, if present.
//
// A name matches whether stored as a blob-like or as a subtree entry,
// so both orderings are searched.
// The returned entry is a copy; mutating it does not affect the tree.
func (tree *Tree) Find(name string) (Entry, bool) {
	for _, searchIsTree := range [...]bool{true, false} {
		index, ok := slices.BinarySearchFunc(tree.entries, name, func(existing Entry, target string) int {
			return nameCompare(existing.Name, existing.Mode == mode.Directory, target, searchIsTree)
		})
		if ok && tree.entries[index].Name == name {
			return tree.entries[index], true
		}
	}

	return Entry{}, false
}
