package tree

import (
	"bytes"
	"fmt"
)

// RemoveEntry removes a tree entry by name.
func (tree *Tree) RemoveEntry(name []byte) error {
	if len(tree.Entries) == 0 {
		return fmt.Errorf("object: tree: entry %q not found", name)
	}

	for i := range tree.Entries {
		if bytes.Equal(tree.Entries[i].Name, name) {
			copy(tree.Entries[i:], tree.Entries[i+1:])
			tree.Entries = tree.Entries[:len(tree.Entries)-1]

			return nil
		}
	}

	return fmt.Errorf("object: tree: entry %q not found", name)
}
