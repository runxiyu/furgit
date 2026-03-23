package resolve

import (
	"fmt"
	"io/fs"
	"strings"

	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
)

func (treeFS *TreeFS) resolvePath(op treeFSOp, name string) (treeEntryValue, error) {
	if !treeFSValidPath(name) {
		return treeEntryValue{}, treeFSPathError(op, name, fs.ErrInvalid)
	}

	if name == "." {
		return treeEntryValue{
			name:      ".",
			mode:      object.FileModeDir,
			treeID:    treeFS.rootTree,
			treeEntry: treeFS.rootEntry,
		}, nil
	}

	entry, err := treeFS.resolver.Path(treeFS.rootTree, treeFSSplitPath(name))
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
	if err != nil && strings.Contains(err.Error(), "not found") {
		return treeFSPathError(op, name, fs.ErrNotExist)
	}

	if err != nil && strings.Contains(err.Error(), "is not a tree") {
		return treeFSPathError(op, name, fs.ErrInvalid)
	}

	return treeFSPathError(op, name, err)
}

type treeEntryValue struct {
	name      string
	mode      object.FileMode
	objectID  objectid.ObjectID
	treeID    objectid.ObjectID
	treeEntry *object.TreeEntry
}

func (entry treeEntryValue) isDir() bool {
	return entry.mode == object.FileModeDir
}

func (entry treeEntryValue) blobSize(resolve *Resolver) (int64, error) {
	_, size, err := resolve.store.ReadHeader(entry.objectID)
	if err != nil {
		return 0, err
	}

	return size, nil
}

func (entry treeEntryValue) subtreeID() (objectid.ObjectID, error) {
	if entry.name == "." {
		return entry.treeID, nil
	}

	if entry.mode != object.FileModeDir {
		return objectid.ObjectID{}, fmt.Errorf("object/resolve: path %q is not a tree", entry.name)
	}

	return entry.objectID, nil
}
