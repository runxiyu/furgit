package resolve

import (
	"fmt"
	"io"

	"codeberg.org/lindenii/furgit/object"
)

func (treeFS *TreeFS) ReadFile(name string) ([]byte, error) {
	entry, err := treeFS.resolvePath(treeFSOpReadFile, name)
	if err != nil {
		return nil, err
	}

	if entry.isDir() {
		return nil, treeFSPathError(treeFSOpReadFile, name, fmt.Errorf("is a directory"))
	}

	if entry.mode == object.FileModeGitlink {
		return nil, treeFSPathError(treeFSOpReadFile, name, fmt.Errorf("object/resolve: gitlink entries are not readable as files"))
	}

	reader, _, err := treeFS.resolver.ExactBlobReader(entry.objectID)
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
