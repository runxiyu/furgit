package resolve

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
)

// Path resolves parts within the tree identified by root and returns the final
// tree entry.
//
// The root object may be any tree-ish object accepted by PeelToTree.
//
// parts must contain at least one path segment. Intermediate path segments
// must resolve to tree entries. The final entry is returned without loading
// its object.
func (r *Resolver) Path(root objectid.ObjectID, parts [][]byte) (object.TreeEntry, error) {
	if len(parts) == 0 {
		return object.TreeEntry{}, fmt.Errorf("object/resolve: empty tree path")
	}

	current, err := r.PeelToTree(root)
	if err != nil {
		return object.TreeEntry{}, err
	}

	for i, part := range parts {
		if len(part) == 0 {
			return object.TreeEntry{}, fmt.Errorf("object/resolve: empty tree path segment")
		}

		entry := current.Object().Entry(part)
		if entry == nil {
			return object.TreeEntry{}, fmt.Errorf("object/resolve: tree entry %q not found", part)
		}

		if i == len(parts)-1 {
			return *entry, nil
		}

		if entry.Mode != object.FileModeDir {
			return object.TreeEntry{}, fmt.Errorf("object/resolve: path segment %q is not a tree", part)
		}

		current, err = r.ExactTree(entry.ID)
		if err != nil {
			return object.TreeEntry{}, err
		}
	}

	return object.TreeEntry{}, fmt.Errorf("object/resolve: tree entry not found")
}
