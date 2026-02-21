// Package trees provides recursive diffs between Git tree objects.
package trees

import (
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
)

// Diff compares two trees and returns recursive differences.
//
// readTree is used to lazily load child trees by object ID when recursion
// reaches directory entries.
func Diff(a, b *object.Tree, readTree func(objectid.ObjectID) (*object.Tree, error)) ([]Entry, error) {
	var out []Entry
	if err := diffRecursive(a, b, nil, readTree, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func diffRecursive(a, b *object.Tree, prefix []byte, readTree func(objectid.ObjectID) (*object.Tree, error), out *[]Entry) error {
	if a == nil && b == nil {
		return nil
	}

	if a == nil {
		for i := range b.Entries {
			entry := &b.Entries[i]
			full := joinPath(prefix, entry.Name)
			*out = append(*out, Entry{Path: full, Kind: EntryKindAdded, Old: nil, New: entry})
			if entry.Mode == object.FileModeDir {
				sub, err := readTree(entry.ID)
				if err != nil {
					return err
				}
				if err := diffRecursive(nil, sub, full, readTree, out); err != nil {
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
			*out = append(*out, Entry{Path: full, Kind: EntryKindDeleted, Old: entry, New: nil})
			if entry.Mode == object.FileModeDir {
				sub, err := readTree(entry.ID)
				if err != nil {
					return err
				}
				if err := diffRecursive(sub, nil, full, readTree, out); err != nil {
					return err
				}
			}
		}
		return nil
	}

	i := 0
	j := 0
	for i < len(a.Entries) && j < len(b.Entries) {
		left := &a.Entries[i]
		right := &b.Entries[j]
		cmp := object.TreeEntryNameCompare(
			left.Name,
			left.Mode,
			right.Name,
			right.Mode == object.FileModeDir,
		)
		switch {
		case cmp < 0:
			full := joinPath(prefix, left.Name)
			*out = append(*out, Entry{Path: full, Kind: EntryKindDeleted, Old: left, New: nil})
			if left.Mode == object.FileModeDir {
				sub, err := readTree(left.ID)
				if err != nil {
					return err
				}
				if err := diffRecursive(sub, nil, full, readTree, out); err != nil {
					return err
				}
			}
			i++
		case cmp > 0:
			full := joinPath(prefix, right.Name)
			*out = append(*out, Entry{Path: full, Kind: EntryKindAdded, Old: nil, New: right})
			if right.Mode == object.FileModeDir {
				sub, err := readTree(right.ID)
				if err != nil {
					return err
				}
				if err := diffRecursive(nil, sub, full, readTree, out); err != nil {
					return err
				}
			}
			j++
		default:
			full := joinPath(prefix, left.Name)
			modified := left.Mode != right.Mode || left.ID != right.ID
			if modified {
				*out = append(*out, Entry{Path: full, Kind: EntryKindModified, Old: left, New: right})
			}
			if left.Mode == object.FileModeDir && right.Mode == object.FileModeDir && left.ID != right.ID {
				leftSub, err := readTree(left.ID)
				if err != nil {
					return err
				}
				rightSub, err := readTree(right.ID)
				if err != nil {
					return err
				}
				if err := diffRecursive(leftSub, rightSub, full, readTree, out); err != nil {
					return err
				}
			}
			i++
			j++
		}
	}

	for ; i < len(a.Entries); i++ {
		left := &a.Entries[i]
		full := joinPath(prefix, left.Name)
		*out = append(*out, Entry{Path: full, Kind: EntryKindDeleted, Old: left, New: nil})
		if left.Mode == object.FileModeDir {
			sub, err := readTree(left.ID)
			if err != nil {
				return err
			}
			if err := diffRecursive(sub, nil, full, readTree, out); err != nil {
				return err
			}
		}
	}

	for ; j < len(b.Entries); j++ {
		right := &b.Entries[j]
		full := joinPath(prefix, right.Name)
		*out = append(*out, Entry{Path: full, Kind: EntryKindAdded, Old: nil, New: right})
		if right.Mode == object.FileModeDir {
			sub, err := readTree(right.ID)
			if err != nil {
				return err
			}
			if err := diffRecursive(nil, sub, full, readTree, out); err != nil {
				return err
			}
		}
	}

	return nil
}
