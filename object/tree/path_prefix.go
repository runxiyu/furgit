package tree

import (
	"bytes"
	"slices"
)

// HasPathPrefix reports whether path begins with prefix as whole components.
func HasPathPrefix(path, prefix [][]byte) bool {
	if len(prefix) == 0 {
		return true
	}

	if len(path) < len(prefix) {
		return false
	}

	return slices.EqualFunc(path[:len(prefix)], prefix, bytes.Equal)
}
