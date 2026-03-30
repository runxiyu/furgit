package tree

import (
	"fmt"
	"slices"
)

// InsertEntry inserts a tree entry while preserving Git ordering.
//
// InsertEntry copies newEntry.Name.
func (tree *Tree) InsertEntry(newEntry TreeEntry) error {
	if tree.entry(newEntry.Name, true) != nil || tree.entry(newEntry.Name, false) != nil {
		return fmt.Errorf("object: tree: entry %q already exists", newEntry.Name)
	}

	newEntry.Name = append([]byte(nil), newEntry.Name...)

	insertAt, _ := slices.BinarySearchFunc(tree.Entries, newEntry.Name, func(entry TreeEntry, name []byte) int {
		return TreeEntryNameCompare(entry.Name, entry.Mode, name, newEntry.Mode == FileModeDir)
	})
	tree.Entries = slices.Insert(tree.Entries, insertAt, newEntry)

	return nil
}
