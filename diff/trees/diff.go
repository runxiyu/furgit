// Package trees provides recursive diffs between Git tree objects.
package trees

import (
	objectid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/tree"
)

// Diff compares two trees and returns recursive differences.
//
// readTree is used to lazily load child trees by object ID when recursion
// reaches directory entries.
func Diff(a, b *tree.Tree, readTree func(objectid.ObjectID) (*tree.Tree, error)) ([]Entry, error) {
	var out []Entry

	err := diffRecursive(a, b, nil, readTree, &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}
