package fetch

import (
	"errors"
	"fmt"
	"io/fs"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/tree"
)

func (treeFS *TreeFS) resolvePath(op treeFSOp, name string) (treeEntryValue, error) {
	if !treeFSValidPath(name) {
		return treeEntryValue{}, treeFSPathError(op, name, fs.ErrInvalid)
	}

	if name == "." {
		return treeEntryValue{
			name:      ".",
			mode:      tree.FileModeDir,
			treeID:    treeFS.rootTree,
			treeEntry: treeFS.rootEntry,
		}, nil
	}

	entry, err := treeFS.fetcher.Path(treeFS.rootTree, tree.SplitPath([]byte(name)))
	if err != nil {
		return treeEntryValue{}, treeFS.pathResolveError(op, name, err)
	}

	return treeEntryValue{
		name:      string(entry.Name),
		mode:      entry.Mode,
		objectID:  entry.ID,
		treeEntry: &entry,
	}, nil
}

func (treeFS *TreeFS) pathResolveError(op treeFSOp, name string, err error) error {
	if _, ok := errors.AsType[*PathNotFoundError](err); ok {
		return treeFSPathError(op, name, fs.ErrNotExist)
	}

	if _, ok := errors.AsType[*PathNotTreeError](err); ok {
		return treeFSPathError(op, name, fs.ErrInvalid)
	}

	if _, ok := errors.AsType[*PathEmptyError](err); ok {
		return treeFSPathError(op, name, fs.ErrInvalid)
	}

	if _, ok := errors.AsType[*PathSegmentEmptyError](err); ok {
		return treeFSPathError(op, name, fs.ErrInvalid)
	}

	return treeFSPathError(op, name, err)
}

type treeEntryValue struct {
	name      string
	mode      tree.FileMode
	objectID  objectid.ObjectID
	treeID    objectid.ObjectID
	treeEntry *tree.TreeEntry
}

func (entry treeEntryValue) isDir() bool {
	return entry.mode == tree.FileModeDir
}

func (entry treeEntryValue) blobSize(fetcher *Fetcher) (int64, error) {
	return fetcher.Size(entry.objectID)
}

func (entry treeEntryValue) subtreeID() (objectid.ObjectID, error) {
	if entry.name == "." {
		return entry.treeID, nil
	}

	if entry.mode != tree.FileModeDir {
		return objectid.ObjectID{}, fmt.Errorf("object/fetch: path %q is not a tree", entry.name)
	}

	return entry.objectID, nil
}
