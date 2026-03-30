package tree

import (
	"fmt"
	"slices"
)

// RemoveEntry removes a tree entry by name.
func (tree *Tree) RemoveEntry(name []byte) error {
	if len(tree.Entries) == 0 {
		return fmt.Errorf("object: tree: entry %q not found", name)
	}

	index, ok := tree.entryIndex(name)
	if !ok {
		return fmt.Errorf("object: tree: entry %q not found", name)
	}

	tree.Entries = slices.Delete(tree.Entries, index, index+1)

	return nil
}
