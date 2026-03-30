package tree

import (
	"bytes"
	"fmt"
	"slices"
)

// RemoveEntry removes a tree entry by name.
func (tree *Tree) RemoveEntry(name []byte) error {
	if len(tree.Entries) == 0 {
		return fmt.Errorf("object: tree: entry %q not found", name)
	}

	index := slices.IndexFunc(tree.Entries, func(entry TreeEntry) bool {
		return bytes.Equal(entry.Name, name)
	})
	if index >= 0 {
		tree.Entries = slices.Delete(tree.Entries, index, index+1)
		return nil
	}

	return fmt.Errorf("object: tree: entry %q not found", name)
}
