package resolve

import (
	"fmt"
	"io"
	"io/fs"

	"codeberg.org/lindenii/furgit/object"
)

// Open opens name for reading.
//
// Directories are returned as fs.ReadDirFile values. Gitlink entries are not
// readable through TreeFS.
func (treeFS *TreeFS) Open(name string) (fs.File, error) {
	entry, err := treeFS.resolvePath(treeFSOpOpen, name)
	if err != nil {
		return nil, err
	}

	info, err := treeFS.statEntry(entry)
	if err != nil {
		return nil, treeFSPathError(treeFSOpOpen, name, err)
	}

	if entry.isDir() {
		treeID, err := entry.subtreeID()
		if err != nil {
			return nil, treeFSPathError(treeFSOpOpen, name, err)
		}

		tree, err := treeFS.resolver.ExactTree(treeID)
		if err != nil {
			return nil, treeFSPathError(treeFSOpOpen, name, err)
		}

		entries := make([]fs.DirEntry, 0, len(tree.Object().Entries))
		for _, child := range tree.Object().Entries {
			childEntry := treeEntryValue{
				name:      string(child.Name),
				mode:      child.Mode,
				objectID:  child.ID,
				treeEntry: &child,
			}

			childInfo, err := treeFS.statEntry(childEntry)
			if err != nil {
				return nil, treeFSPathError(treeFSOpOpen, name, err)
			}

			entries = append(entries, childInfo)
		}

		return &treeFSDir{
			info:    info,
			entries: entries,
		}, nil
	}

	if entry.mode == object.FileModeGitlink {
		return nil, treeFSPathError(treeFSOpOpen, name, fmt.Errorf("object/resolve: gitlink entries are not readable as files"))
	}

	reader, _, err := treeFS.resolver.ExactBlobReader(entry.objectID)
	if err != nil {
		return nil, treeFSPathError(treeFSOpOpen, name, err)
	}

	return &treeFSBlob{
		info:   info,
		reader: reader,
	}, nil
}

type treeFSBlob struct {
	info   *treeFSInfo
	reader io.ReadCloser
}

var _ fs.File = (*treeFSBlob)(nil)

func (file *treeFSBlob) Stat() (fs.FileInfo, error) { return file.info, nil }
func (file *treeFSBlob) Read(p []byte) (int, error) { return file.reader.Read(p) }
func (file *treeFSBlob) Close() error               { return file.reader.Close() }

type treeFSDir struct {
	info    *treeFSInfo
	entries []fs.DirEntry
	offset  int
}

var (
	_ fs.File        = (*treeFSDir)(nil)
	_ fs.ReadDirFile = (*treeFSDir)(nil)
)

func (dir *treeFSDir) Stat() (fs.FileInfo, error) { return dir.info, nil }
func (dir *treeFSDir) Close() error               { return nil }

func (dir *treeFSDir) Read(_ []byte) (int, error) {
	return 0, fs.ErrInvalid
}

func (dir *treeFSDir) ReadDir(n int) ([]fs.DirEntry, error) {
	if dir.offset >= len(dir.entries) && n > 0 {
		return nil, io.EOF
	}

	if n <= 0 {
		out := append([]fs.DirEntry(nil), dir.entries[dir.offset:]...)
		dir.offset = len(dir.entries)

		return out, nil
	}

	end := dir.offset + n
	if end > len(dir.entries) {
		end = len(dir.entries)
	}

	out := append([]fs.DirEntry(nil), dir.entries[dir.offset:end]...)
	dir.offset = end

	return out, nil
}
