package tree

import (
	"bytes"

	objectid "codeberg.org/lindenii/furgit/object/id"
)

// TreeEntry represents a single entry in a tree.
type TreeEntry struct {
	Mode FileMode
	Name []byte
	ID   objectid.ObjectID
}

func (tree *Tree) entry(name []byte, searchIsTree bool) *TreeEntry {
	low, high := 0, len(tree.Entries)-1
	for low <= high {
		mid := low + (high-low)/2
		entry := &tree.Entries[mid]

		cmp := TreeEntryNameCompare(entry.Name, entry.Mode, name, searchIsTree)
		if cmp == 0 {
			if bytes.Equal(entry.Name, name) {
				return entry
			}

			return nil
		}

		if cmp < 0 {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	return nil
}
