package resolve

import "io/fs"

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
