package tree

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"lindenii.org/go/furgit/object/tree/mode"
)

// ErrInvalidTree indicates a malformed tree object or tree entry.
var ErrInvalidTree = errors.New("object/tree: invalid tree")

// Insert adds an entry to the tree while preserving Git tree ordering.
//
// It rejects entries with an invalid name or mode,
// and entries whose name conflicts with one already present.
func (tree *Tree) Insert(entry Entry) error {
	if err := validateName(entry.Name); err != nil {
		return err
	}

	if !entry.Mode.IsValid() {
		return fmt.Errorf("%w: entry %q has invalid mode", ErrInvalidTree, entry.Name)
	}

	if _, found := tree.find(entry.Name); found {
		return fmt.Errorf("%w: entry %q already exists", ErrInvalidTree, entry.Name)
	}

	isTree := entry.Mode == mode.Directory

	insertAt, _ := slices.BinarySearchFunc(tree.entries, entry, func(existing Entry, target Entry) int {
		return nameCompare(existing.Name, existing.Mode == mode.Directory, target.Name, isTree)
	})

	tree.entries = slices.Insert(tree.entries, insertAt, entry)

	return nil
}

// find returns the index of the entry with the given name, if present.
//
// A name conflicts whether stored as a blob-like or as a subtree entry,
// so both orderings are searched.
func (tree *Tree) find(name string) (int, bool) {
	for _, searchIsTree := range [...]bool{true, false} {
		index, ok := slices.BinarySearchFunc(tree.entries, name, func(existing Entry, target string) int {
			return nameCompare(existing.Name, existing.Mode == mode.Directory, target, searchIsTree)
		})
		if ok && tree.entries[index].Name == name {
			return index, true
		}
	}

	return 0, false
}

// validateName checks that name is a structurally valid tree entry name.
func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: empty entry name", ErrInvalidTree)
	}

	if strings.IndexByte(name, 0) >= 0 {
		return fmt.Errorf("%w: entry name %q contains NUL", ErrInvalidTree, name)
	}

	if strings.IndexByte(name, '/') >= 0 {
		return fmt.Errorf("%w: entry name %q contains '/'", ErrInvalidTree, name)
	}

	return nil
}
