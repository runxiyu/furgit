package tree

import (
	"fmt"
	"sort"
)

// InsertEntry inserts a tree entry while preserving Git ordering.
func (tree *Tree) InsertEntry(newEntry TreeEntry) error {
	if tree.entry(newEntry.Name, true) != nil || tree.entry(newEntry.Name, false) != nil {
		return fmt.Errorf("object: tree: entry %q already exists", newEntry.Name)
	}

	newIsTree := newEntry.Mode == FileModeDir
	insertAt := sort.Search(len(tree.Entries), func(i int) bool {
		return TreeEntryNameCompare(tree.Entries[i].Name, tree.Entries[i].Mode, newEntry.Name, newIsTree) >= 0
	})
	tree.Entries = append(tree.Entries, TreeEntry{})
	copy(tree.Entries[insertAt+1:], tree.Entries[insertAt:])
	tree.Entries[insertAt] = newEntry

	return nil
}
