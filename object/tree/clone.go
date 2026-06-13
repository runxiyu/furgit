package tree

import "bytes"

// Clone returns a deep copy of the tree
// whose entry names are independent of any memory the original may alias.
//
// Labels: Life-Independent.
func (tree *Tree) Clone() *Tree {
	if tree.entries == nil {
		return &Tree{}
	}

	clone := &Tree{entries: make([]Entry, len(tree.entries))}
	for i, entry := range tree.entries {
		clone.entries[i] = Entry{
			Mode: entry.Mode,
			Name: bytes.Clone(entry.Name),
			ID:   entry.ID,
		}
	}

	return clone
}
