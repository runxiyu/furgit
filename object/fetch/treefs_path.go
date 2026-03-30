package fetch

import "io/fs"

func treeFSValidPath(name string) bool {
	return name == "." || fs.ValidPath(name)
}

func treeFSPathError(op treeFSOp, path string, err error) error {
	return &fs.PathError{Op: op.pathErrorOp(), Path: path, Err: err}
}
