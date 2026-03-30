package tree

import (
	"bytes"
)

// SplitPath splits one slash-separated tree path into components.
func SplitPath(path []byte) [][]byte {
	if len(path) == 0 {
		return nil
	}

	parts := bytes.Split(path, []byte{'/'})
	for i := range parts {
		parts[i] = bytes.Clone(parts[i])
	}

	return parts
}
