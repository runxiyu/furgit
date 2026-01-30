package furgit

// TreeDiffEntryKind represents the type of difference between two tree entries.
type TreeDiffEntryKind int

const (
	// TreeDiffEntryKindInvalid indicates an invalid difference type.
	TreeDiffEntryKindInvalid TreeDiffEntryKind = iota
	// TreeDiffEntryKindDeleted indicates that the entry was deleted.
	TreeDiffEntryKindDeleted
	// TreeDiffEntryKindAdded indicates that the entry was added.
	TreeDiffEntryKindAdded
	// TreeDiffEntryKindModified indicates that the entry was modified.
	TreeDiffEntryKindModified
)

// TreeDiffEntry represents a difference between two tree entries.
type TreeDiffEntry struct {
	// Path is the full slash-separated path relative to the root
	// of the repository.
	Path []byte
	// Kind indicates the type of difference.
	Kind TreeDiffEntryKind
	// Old is the old tree entry (nil iff added).
	Old *TreeEntry
	// New is the new tree entry (nil iff deleted).
	New *TreeEntry
}

// DiffTrees compares two trees rooted at a and b and returns all differences
// as a flat slice of TreeDiffEntry. Differences are discovered recursively.
func (repo *Repository) DiffTrees(a, b *StoredTree) ([]TreeDiffEntry, error) {
	var out []TreeDiffEntry
	err := repo.diffTreesRecursive(a, b, nil, &out)
	return out, err
}

func (repo *Repository) diffTreesRecursive(a, b *StoredTree, prefix []byte, out *[]TreeDiffEntry) error {
	if a == nil && b == nil {
		return nil
	}

	if a == nil {
		for i := range b.Entries {
			entry := &b.Entries[i]
			full := joinPath(prefix, entry.Name)

			*out = append(*out, TreeDiffEntry{
				Path: full,
				Kind: TreeDiffEntryKindAdded,
				Old:  nil,
				New:  entry,
			})

			if entry.Mode == FileModeDir {
				sub, err := repo.readTree(entry.ID)
				if err != nil {
					return err
				}
				if err := repo.diffTreesRecursive(nil, sub, full, out); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if b == nil {
		for i := range a.Entries {
			entry := &a.Entries[i]
			full := joinPath(prefix, entry.Name)

			*out = append(*out, TreeDiffEntry{
				Path: full,
				Kind: TreeDiffEntryKindDeleted,
				Old:  entry,
				New:  nil,
			})

			if entry.Mode == FileModeDir {
				sub, err := repo.readTree(entry.ID)
				if err != nil {
					return err
				}
				if err := repo.diffTreesRecursive(sub, nil, full, out); err != nil {
					return err
				}
			}
		}
		return nil
	}

	left := make(map[string]*TreeEntry, len(a.Entries))
	for i := range a.Entries {
		e := &a.Entries[i]
		left[string(e.Name)] = e
	}
	right := make(map[string]*TreeEntry, len(b.Entries))
	for i := range b.Entries {
		e := &b.Entries[i]
		right[string(e.Name)] = e
	}

	seen := make(map[string]bool, len(a.Entries)+len(b.Entries))
	for n := range left {
		seen[n] = true
	}
	for n := range right {
		seen[n] = true
	}

	for name := range seen {
		le := left[name]
		re := right[name]

		full := joinPath(prefix, []byte(name))

		switch {
		case le == nil && re != nil:
			*out = append(*out, TreeDiffEntry{
				Path: full,
				Kind: TreeDiffEntryKindAdded,
				Old:  nil,
				New:  re,
			})

			if re.Mode == FileModeDir {
				sub, err := repo.readTree(re.ID)
				if err != nil {
					return err
				}
				if err := repo.diffTreesRecursive(nil, sub, full, out); err != nil {
					return err
				}
			}

		case le != nil && re == nil:
			*out = append(*out, TreeDiffEntry{
				Path: full,
				Kind: TreeDiffEntryKindDeleted,
				Old:  le,
				New:  nil,
			})

			if le.Mode == FileModeDir {
				sub, err := repo.readTree(le.ID)
				if err != nil {
					return err
				}
				if err := repo.diffTreesRecursive(sub, nil, full, out); err != nil {
					return err
				}
			}

		default:
			modified := (le.Mode != re.Mode) || (le.ID != re.ID)
			if modified {
				*out = append(*out, TreeDiffEntry{
					Path: full,
					Kind: TreeDiffEntryKindModified,
					Old:  le,
					New:  re,
				})
			}

			if le.Mode == FileModeDir && re.Mode == FileModeDir && le.ID != re.ID {
				ls, err := repo.readTree(le.ID)
				if err != nil {
					return err
				}
				rs, err := repo.readTree(re.ID)
				if err != nil {
					return err
				}
				if err := repo.diffTreesRecursive(ls, rs, full, out); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func joinPath(prefix, name []byte) []byte {
	if len(prefix) == 0 {
		out := make([]byte, len(name))
		copy(out, name)
		return out
	}
	out := make([]byte, len(prefix)+1+len(name))
	copy(out, prefix)
	out[len(prefix)] = '/'
	copy(out[len(prefix)+1:], name)
	return out
}

func (repo *Repository) readTree(id Hash) (*StoredTree, error) {
	obj, err := repo.ReadObject(id)
	if err != nil {
		return nil, err
	}
	tree, ok := obj.(*StoredTree)
	if !ok {
		return nil, ErrInvalidObject
	}
	return tree, nil
}
