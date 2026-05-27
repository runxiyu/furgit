package tree

import (
	"bytes"
	"slices"

	objectid "lindenii.org/go/furgit/object/id"
)

// TreeEntry represents a single entry in a tree.
type TreeEntry struct {
	Mode FileMode
	// Name is part of the tree ordering. Mutating it after insertion may break
	// Tree ordering and lookup behavior.
	Name []byte
	ID   objectid.ObjectID
}

func (tree *Tree) entry(name []byte, searchIsTree bool) *TreeEntry {
	index, ok := slices.BinarySearchFunc(tree.Entries, name, func(entry TreeEntry, name []byte) int {
		return TreeEntryNameCompare(entry.Name, entry.Mode, name, searchIsTree)
	})
	if !ok {
		return nil
	}

	entry := &tree.Entries[index]
	if !bytes.Equal(entry.Name, name) {
		return nil
	}

	return entry
}

func (tree *Tree) entryIndex(name []byte) (int, bool) {
	index, ok := tree.entryIndexWithMode(name, true)
	if ok {
		return index, true
	}

	return tree.entryIndexWithMode(name, false)
}

func (tree *Tree) entryIndexWithMode(name []byte, searchIsTree bool) (int, bool) {
	index, ok := slices.BinarySearchFunc(tree.Entries, name, func(entry TreeEntry, name []byte) int {
		return TreeEntryNameCompare(entry.Name, entry.Mode, name, searchIsTree)
	})
	if !ok {
		return 0, false
	}

	if !bytes.Equal(tree.Entries[index].Name, name) {
		return 0, false
	}

	return index, true
}
