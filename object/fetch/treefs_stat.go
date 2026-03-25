package fetch

import "io/fs"

// Stat returns synthetic file metadata for name.
//
// TreeFS metadata reflects Git tree entry mode and blob size where applicable.
// It does not represent filesystem stat metadata: ModTime is zero, ownership is
// unavailable, and Sys returns the underlying tree.TreeEntry when one exists.
func (treeFS *TreeFS) Stat(name string) (fs.FileInfo, error) {
	entry, err := treeFS.resolvePath(treeFSOpStat, name)
	if err != nil {
		return nil, err
	}

	info, err := treeFS.statEntry(entry)
	if err != nil {
		return nil, treeFSPathError(treeFSOpStat, name, err)
	}

	return info, nil
}
