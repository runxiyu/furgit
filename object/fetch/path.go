package fetch

import (
	"errors"
	"fmt"

	oid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/tree"
	"lindenii.org/go/furgit/object/tree/mode"
)

var ErrPathInvalid = errors.New("object/fetch: invalid tree path")

// PathNotFoundError indicates that one tree path segment was not found.
type PathNotFoundError struct {
	Index int
	Name  []byte
}

func (err *PathNotFoundError) Error() string {
	return fmt.Sprintf("object/fetch: tree entry %q not found at index %d", err.Name, err.Index)
}

// PathNotTreeError indicates that one intermediate path segment was not a tree.
type PathNotTreeError struct {
	Index int
	Name  []byte
}

func (err *PathNotTreeError) Error() string {
	return fmt.Sprintf("object/fetch: path segment %q at index %d is not a tree", err.Name, err.Index)
}

// Path resolves parts within the tree identified by root
// and returns the final tree entry.
//
// The root object may be any tree-ish object accepted by PeelToTree.
//
// parts must contain at least one path segment.
// Intermediate path segments must resolve to tree entries.
// The final entry is returned without loading its object.
// Path segments may not contain \x00.
//
// If your entry names are valid UTF-8
// and uses / solely as segment separators,
// it may be convenient to use TreeFS
// for an io/fs.FS-like interface.
//
// Labels: Life-Parent.
func (fetcher *Fetcher) Path(root oid.ObjectID, parts [][]byte) (tree.Entry, error) {
	if len(parts) == 0 {
		return tree.Entry{}, ErrPathInvalid
	}

	current, err := fetcher.PeelToTree(root)
	if err != nil {
		return tree.Entry{}, err
	}

	for i, part := range parts {
		if len(part) == 0 {
			return tree.Entry{}, ErrPathInvalid
		}

		entry, ok := current.Object().Find(part)
		if !ok {
			return tree.Entry{}, &PathNotFoundError{
				Index: i,
				Name:  append([]byte(nil), part...),
			}
		}

		if i == len(parts)-1 {
			return entry, nil
		}

		if entry.Mode != mode.Directory {
			return tree.Entry{}, &PathNotTreeError{
				Index: i,
				Name:  append([]byte(nil), part...),
			}
		}

		current, err = fetcher.ExactTree(entry.ID)
		if err != nil {
			return tree.Entry{}, err
		}
	}

	panic("object/fetch: unreachable")
}
