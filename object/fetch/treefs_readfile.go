package fetch

import (
	"fmt"
	"io"

	"codeberg.org/lindenii/furgit/object/tree"
)

// ReadFile reads the blob contents at name.
//
// Directories and gitlink entries are not readable through TreeFS.
func (treeFS *TreeFS) ReadFile(name string) ([]byte, error) {
	entry, err := treeFS.resolvePath(treeFSOpReadFile, name)
	if err != nil {
		return nil, err
	}

	if entry.isDir() {
		return nil, treeFSPathError(treeFSOpReadFile, name, fmt.Errorf("is a directory"))
	}

	if entry.mode == tree.FileModeGitlink {
		return nil, treeFSPathError(treeFSOpReadFile, name, fmt.Errorf("object/fetch: gitlink entries are not readable as files"))
	}

	reader, _, err := treeFS.fetcher.ExactBlobReader(entry.objectID)
	if err != nil {
		return nil, treeFSPathError(treeFSOpReadFile, name, err)
	}

	defer func() { _ = reader.Close() }()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, treeFSPathError(treeFSOpReadFile, name, err)
	}

	return data, nil
}
