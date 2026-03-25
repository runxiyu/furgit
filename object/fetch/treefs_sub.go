package fetch

import "io/fs"

// Sub returns a new TreeFS rooted at dir.
func (treeFS *TreeFS) Sub(dir string) (fs.FS, error) {
	entry, err := treeFS.resolvePath(treeFSOpSub, dir)
	if err != nil {
		return nil, err
	}

	treeID, err := entry.subtreeID()
	if err != nil {
		return nil, treeFSPathError(treeFSOpSub, dir, fs.ErrInvalid)
	}

	return &TreeFS{
		resolver:  treeFS.resolver,
		rootTree:  treeID,
		rootEntry: entry.treeEntry,
	}, nil
}
