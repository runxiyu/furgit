package resolve

import "io/fs"

func (treeFS *TreeFS) ReadDir(name string) ([]fs.DirEntry, error) {
	file, err := treeFS.Open(name)
	if err != nil {
		return nil, err
	}

	defer func() { _ = file.Close() }()

	readDirFile, ok := file.(fs.ReadDirFile)
	if !ok {
		return nil, treeFSPathError(treeFSOpReadDir, name, fs.ErrInvalid)
	}

	return readDirFile.ReadDir(-1)
}
