package repository

import (
	"errors"
	"fmt"

	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectstored"
)

// ResolveTreeEntry resolves one path within a stored root tree.
//
// parts must contain at least one path segment. Intermediate segments must be
// tree entries.
func (repo *Repository) ResolveTreeEntry(tree *objectstored.StoredTree, parts [][]byte) (object.TreeEntry, error) {
	if tree == nil {
		return object.TreeEntry{}, errors.New("repository: nil root tree")
	}

	if len(parts) == 0 {
		return object.TreeEntry{}, errors.New("repository: empty tree path")
	}

	current := tree

	for i, part := range parts {
		if len(part) == 0 {
			return object.TreeEntry{}, errors.New("repository: empty tree path segment")
		}

		entry := current.Tree().Entry(part)
		if entry == nil {
			return object.TreeEntry{}, fmt.Errorf("repository: tree entry %q not found", part)
		}

		if i == len(parts)-1 {
			return *entry, nil
		}

		if entry.Mode != object.FileModeDir {
			return object.TreeEntry{}, fmt.Errorf("repository: path segment %q is not a tree", part)
		}

		next, err := repo.ReadStoredTree(entry.ID)
		if err != nil {
			return object.TreeEntry{}, err
		}

		current = next
	}

	return object.TreeEntry{}, fmt.Errorf("repository: tree entry not found")
}
