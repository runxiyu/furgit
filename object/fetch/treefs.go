package fetch

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"time"

	oid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/tree"
	"lindenii.org/go/furgit/object/tree/mode"
	"lindenii.org/go/lgo/intconv"
)

// TreeFS exposes one Git tree as an fs.FS view backed by a Fetcher.
//
// TreeFS interprets names using io/fs path rules.
// Those rules do not match raw Git tree entry naming exactly:
// names are UTF-8, slash-separated, and must be valid fs.FS paths.
// Tree entries that cannot be represented under those rules
// are not addressable through this API.
//
// Labels: MT-Safe.
type TreeFS struct {
	fetcher   *Fetcher
	rootTree  oid.ObjectID
	rootEntry *tree.Entry
}

var (
	_ fs.FS         = (*TreeFS)(nil)
	_ fs.ReadFileFS = (*TreeFS)(nil)
	_ fs.ReadDirFS  = (*TreeFS)(nil)
	_ fs.StatFS     = (*TreeFS)(nil)
	_ fs.SubFS      = (*TreeFS)(nil)
)

func splitPath(path string) []string {
	if len(path) == 0 {
		return nil
	}

	return strings.Split(path, "/")
}

type treeEntryValue struct {
	name      string
	mode      mode.Mode
	objectID  oid.ObjectID
	treeID    oid.ObjectID
	treeEntry *tree.Entry
}

func (entry treeEntryValue) isDir() bool {
	return entry.mode == mode.Directory
}

func (entry treeEntryValue) blobSize(fetcher *Fetcher) (uint64, error) {
	return fetcher.Size(entry.objectID)
}

func (entry treeEntryValue) subtreeID() (oid.ObjectID, error) {
	if entry.name == "." {
		return entry.treeID, nil
	}

	if entry.mode != mode.Directory {
		return oid.ObjectID{}, fmt.Errorf("object/fetch: path %q is not a tree", entry.name)
	}

	return entry.objectID, nil
}

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

func treeFSEntryMode(mod mode.Mode) fs.FileMode {
	switch mod {
	case mode.Directory:
		return fs.ModeDir | 0o555
	case mode.Regular:
		return 0o444
	case mode.Executable:
		return 0o555
	case mode.Symlink:
		return fs.ModeSymlink | 0o444
	case mode.Gitlink:
		return fs.ModeIrregular
	default:
		return fs.ModeIrregular
	}
}

// TreeFS returns a new filesystem view rooted at root, which may be any
// tree-ish object accepted by PeelToTreeID.
//
// Labels: Deps-Borrowed, Life-Parent.
func (fetcher *Fetcher) TreeFS(root oid.ObjectID) (*TreeFS, error) {
	rootTree, err := fetcher.PeelToTreeID(root)
	if err != nil {
		return nil, err
	}

	return &TreeFS{
		fetcher:  fetcher,
		rootTree: rootTree,
	}, nil
}

type treeFSOp uint8

const (
	treeFSOpOpen treeFSOp = iota
	treeFSOpReadFile
	treeFSOpReadDir
	treeFSOpStat
	treeFSOpSub
)

func (op treeFSOp) pathErrorOp() string {
	switch op {
	case treeFSOpOpen:
		return "open"
	case treeFSOpReadFile:
		return "readfile"
	case treeFSOpReadDir:
		return "readdir"
	case treeFSOpStat:
		return "stat"
	case treeFSOpSub:
		return "sub"
	default:
		return "treefs"
	}
}

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

		tree, err := treeFS.fetcher.ExactTree(treeID)
		if err != nil {
			return nil, treeFSPathError(treeFSOpOpen, name, err)
		}

		entries := make([]fs.DirEntry, 0, len(tree.Object().Entries()))
		for _, child := range tree.Object().Entries() {
			childEntry := treeEntryValue{
				name:      child.Name,
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

	if entry.mode == mode.Gitlink {
		return nil, treeFSPathError(treeFSOpOpen, name, fmt.Errorf("object/fetch: gitlink entries are not readable as files"))
	}

	reader, _, err := treeFS.fetcher.ExactBlobReader(entry.objectID)
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

	end := min(dir.offset+n, len(dir.entries))

	out := append([]fs.DirEntry(nil), dir.entries[dir.offset:end]...)
	dir.offset = end

	return out, nil
}

func treeFSValidPath(name string) bool {
	return name == "." || fs.ValidPath(name)
}

func treeFSPathError(op treeFSOp, path string, err error) error {
	return &fs.PathError{Op: op.pathErrorOp(), Path: path, Err: err}
}

// ReadDir reads and returns all directory entries for name.
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

	if entry.mode == mode.Gitlink {
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

// Stat returns synthetic file metadata for name.
//
// TreeFS metadata reflects Git tree entry mode and blob size where applicable.
// It does not represent filesystem stat metadata: ModTime is zero, ownership is
// unavailable, and Sys returns the underlying tree.Entry when one exists.
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
		fetcher:   treeFS.fetcher,
		rootTree:  treeID,
		rootEntry: entry.treeEntry,
	}, nil
}

func (treeFS *TreeFS) resolvePath(op treeFSOp, name string) (treeEntryValue, error) {
	if !treeFSValidPath(name) {
		return treeEntryValue{}, treeFSPathError(op, name, fs.ErrInvalid)
	}

	if name == "." {
		return treeEntryValue{
			name:      ".",
			mode:      mode.Directory,
			treeID:    treeFS.rootTree,
			treeEntry: treeFS.rootEntry,
		}, nil
	}

	entry, err := treeFS.fetcher.Path(treeFS.rootTree, splitPath(name))
	if err != nil {
		return treeEntryValue{}, treeFS.pathResolveError(op, name, err)
	}

	return treeEntryValue{
		name:      entry.Name,
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

	if errors.Is(err, ErrPathInvalid) {
		return treeFSPathError(op, name, fs.ErrInvalid)
	}

	return treeFSPathError(op, name, err)
}

func (treeFS *TreeFS) statEntry(entry treeEntryValue) (*treeFSInfo, error) {
	size := int64(0)

	if entry.mode == mode.Regular || entry.mode == mode.Executable || entry.mode == mode.Symlink {
		sz, err := entry.blobSize(treeFS.fetcher)
		if err != nil {
			return nil, err
		}

		size, err = intconv.Uint64ToInt64(sz)
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
