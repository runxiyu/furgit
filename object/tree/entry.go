package tree

import (
	"bytes"
	"slices"

	objectid "codeberg.org/lindenii/furgit/object/id"
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
	index, ok := slices.BinarySearchFunc(tree.Entries, treeEntrySearch{
		name:   name,
		isTree: searchIsTree,
	}, func(entry TreeEntry, search treeEntrySearch) int {
		return TreeEntryNameCompare(entry.Name, entry.Mode, search.name, search.isTree)
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

type treeEntrySearch struct {
	name   []byte
	isTree bool
}
