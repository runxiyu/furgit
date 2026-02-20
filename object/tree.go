package object

import (
	"bytes"
	"fmt"
	"sort"

	"codeberg.org/lindenii/furgit/objecttype"
	"codeberg.org/lindenii/furgit/oid"
)

// FileMode represents the mode of a file in a Git tree.
type FileMode uint32

const (
	FileModeDir        FileMode = 0o40000
	FileModeRegular    FileMode = 0o100644
	FileModeExecutable FileMode = 0o100755
	FileModeSymlink    FileMode = 0o120000
	FileModeGitlink    FileMode = 0o160000
)

// TreeEntry represents a single entry in a tree.
type TreeEntry struct {
	Mode FileMode
	Name []byte
	ID   oid.ObjectID
}

// Tree represents a Git tree object.
type Tree struct {
	Entries []TreeEntry
}

// ObjectType returns TypeTree.
func (tree *Tree) ObjectType() objecttype.Type {
	_ = tree
	return objecttype.TypeTree
}

// Entry looks up a tree entry by name.
func (tree *Tree) Entry(name []byte) *TreeEntry {
	if len(tree.Entries) == 0 {
		return nil
	}
	if e := tree.entry(name, true); e != nil {
		return e
	}
	return tree.entry(name, false)
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

// InsertEntry inserts a tree entry while preserving Git ordering.
func (tree *Tree) InsertEntry(newEntry TreeEntry) error {
	if tree == nil {
		return ErrInvalidObject
	}
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

// RemoveEntry removes a tree entry by name.
func (tree *Tree) RemoveEntry(name []byte) error {
	if tree == nil {
		return ErrInvalidObject
	}
	if len(tree.Entries) == 0 {
		return ErrNotFound
	}
	for i := range tree.Entries {
		if bytes.Equal(tree.Entries[i].Name, name) {
			copy(tree.Entries[i:], tree.Entries[i+1:])
			tree.Entries = tree.Entries[:len(tree.Entries)-1]
			return nil
		}
	}
	return ErrNotFound
}

// TreeEntryNameCompare compares names using Git tree ordering rules.
func TreeEntryNameCompare(entryName []byte, entryMode FileMode, searchName []byte, searchIsTree bool) int {
	isEntryTree := entryMode == FileModeDir

	entryLen := len(entryName)
	if isEntryTree {
		entryLen++
	}
	searchLen := len(searchName)
	if searchIsTree {
		searchLen++
	}

	n := entryLen
	if searchLen < n {
		n = searchLen
	}

	for i := 0; i < n; i++ {
		var ec, sc byte
		if i < len(entryName) {
			ec = entryName[i]
		} else {
			ec = '/'
		}
		if i < len(searchName) {
			sc = searchName[i]
		} else {
			sc = '/'
		}
		if ec < sc {
			return -1
		}
		if ec > sc {
			return 1
		}
	}

	if entryLen < searchLen {
		return -1
	}
	if entryLen > searchLen {
		return 1
	}
	return 0
}
