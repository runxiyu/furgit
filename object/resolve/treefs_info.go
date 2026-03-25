package resolve

import (
	"io/fs"
	"time"

	"codeberg.org/lindenii/furgit/object/tree"
)

type treeFSInfo struct {
	name  string
	mode  fs.FileMode
	size  int64
	sys   any
	isDir bool
}

var (
	_ fs.FileInfo = (*treeFSInfo)(nil)
	_ fs.DirEntry = (*treeFSInfo)(nil)
)

func (info *treeFSInfo) Name() string       { return info.name }
func (info *treeFSInfo) Size() int64        { return info.size }
func (info *treeFSInfo) Mode() fs.FileMode  { return info.mode }
func (info *treeFSInfo) Type() fs.FileMode  { return info.mode.Type() }
func (info *treeFSInfo) IsDir() bool        { return info.isDir }
func (info *treeFSInfo) ModTime() time.Time { return time.Time{} }
func (info *treeFSInfo) Sys() any           { return info.sys }
func (info *treeFSInfo) Info() (fs.FileInfo, error) {
	return info, nil
}

func treeFSEntryMode(mode tree.FileMode) fs.FileMode {
	switch mode {
	case tree.FileModeDir:
		return fs.ModeDir | 0o555
	case tree.FileModeRegular:
		return 0o444
	case tree.FileModeExecutable:
		return 0o555
	case tree.FileModeSymlink:
		return fs.ModeSymlink | 0o444
	case tree.FileModeGitlink:
		return fs.ModeIrregular
	default:
		return fs.ModeIrregular
	}
}

func (treeFS *TreeFS) statEntry(entry treeEntryValue) (*treeFSInfo, error) {
	size := int64(0)

	if entry.mode == tree.FileModeRegular || entry.mode == tree.FileModeExecutable || entry.mode == tree.FileModeSymlink {
		var err error

		size, err = entry.blobSize(treeFS.resolver)
		if err != nil {
			return nil, err
		}
	}

	var sys any
	if entry.treeEntry != nil {
		sys = *entry.treeEntry
	}

	return &treeFSInfo{
		name:  entry.name,
		mode:  treeFSEntryMode(entry.mode),
		size:  size,
		sys:   sys,
		isDir: entry.isDir(),
	}, nil
}
