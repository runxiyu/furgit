package tree

import (
	"bytes"
	"slices"
)

// ClonePath returns one deep copy of path.
func ClonePath(path [][]byte) [][]byte {
	cloned := slices.Clone(path)
	for i := range cloned {
		cloned[i] = bytes.Clone(cloned[i])
	}

	return cloned
}
