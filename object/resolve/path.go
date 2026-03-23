package resolve

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
)

// PathEmptyError indicates that Path received no segments.
type PathEmptyError struct{}

func (err *PathEmptyError) Error() string {
	return "object/resolve: empty tree path"
}

// PathSegmentEmptyError indicates that one path segment is empty.
type PathSegmentEmptyError struct {
	Index int
}

func (err *PathSegmentEmptyError) Error() string {
	return fmt.Sprintf("object/resolve: empty tree path segment at index %d", err.Index)
}

// PathNotFoundError indicates that one tree path segment was not found.
type PathNotFoundError struct {
	Index int
	Name  []byte
}

func (err *PathNotFoundError) Error() string {
	return fmt.Sprintf("object/resolve: tree entry %q not found at index %d", err.Name, err.Index)
}

// PathNotTreeError indicates that one intermediate path segment was not a tree.
type PathNotTreeError struct {
	Index int
	Name  []byte
}

func (err *PathNotTreeError) Error() string {
	return fmt.Sprintf("object/resolve: path segment %q at index %d is not a tree", err.Name, err.Index)
}

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
		return object.TreeEntry{}, &PathEmptyError{}
	}

	current, err := r.PeelToTree(root)
	if err != nil {
		return object.TreeEntry{}, err
	}

	for i, part := range parts {
		if len(part) == 0 {
			return object.TreeEntry{}, &PathSegmentEmptyError{Index: i}
		}

		entry := current.Object().Entry(part)
		if entry == nil {
			return object.TreeEntry{}, &PathNotFoundError{
				Index: i,
				Name:  append([]byte(nil), part...),
			}
		}

		if i == len(parts)-1 {
			return *entry, nil
		}

		if entry.Mode != object.FileModeDir {
			return object.TreeEntry{}, &PathNotTreeError{
				Index: i,
				Name:  append([]byte(nil), part...),
			}
		}

		current, err = r.ExactTree(entry.ID)
		if err != nil {
			return object.TreeEntry{}, err
		}
	}

	return object.TreeEntry{}, &PathNotFoundError{Index: len(parts) - 1}
}
