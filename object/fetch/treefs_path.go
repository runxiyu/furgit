package fetch

import (
	"io/fs"
	"strings"
)

func treeFSValidPath(name string) bool {
	return name == "." || fs.ValidPath(name)
}

func treeFSSplitPath(name string) [][]byte {
	if name == "." {
		return nil
	}

	parts := strings.Split(name, "/")

	out := make([][]byte, len(parts))
	for i, part := range parts {
		out[i] = []byte(part)
	}

	return out
}

func treeFSPathError(op treeFSOp, path string, err error) error {
	return &fs.PathError{Op: op.pathErrorOp(), Path: path, Err: err}
}
