package files

import (
	"fmt"
	"os"
	"path"
	"time"

	"lindenii.org/go/furgit/object/id"
	refname "lindenii.org/go/furgit/ref/name"
	"lindenii.org/go/furgit/ref/store"
)

// FsyncMode controls whether written references are flushed
// to stable storage before their commit rename.
type FsyncMode uint8

const (
	// FsyncNone performs no explicit flushing.
	FsyncNone FsyncMode = iota

	// FsyncAlways flushes each written file before its commit rename.
	FsyncAlways
)

// Lock timeout defaults.
const (
	DefaultLooseLockTimeout  = 100 * time.Millisecond
	DefaultPackedLockTimeout = time.Second
)

// Options configures one files reference store.
//
//exhaustruct:optional
type Options struct {
	// LooseLockTimeout bounds waiting for one loose reference lock.
	// Zero means one attempt without waiting;
	// negative means waiting without bound.
	LooseLockTimeout time.Duration

	// PackedLockTimeout bounds waiting for the packed-refs lock.
	// Zero means one attempt without waiting;
	// negative means waiting without bound.
	PackedLockTimeout time.Duration

	// Fsync controls flushing of written references to stable storage.
	Fsync FsyncMode
}

// Files reads and writes one Git files-backend reference namespace.
//
// Labels: MT-Safe, Deps-Borrowed, Life-Parent, Close-Caller.
type Files struct {
	gitRoot      *os.Root
	commonRoot   *os.Root
	objectFormat id.ObjectFormat
	options      Options
}

var (
	_ store.Reader        = (*Files)(nil)
	_ store.Transactioner = (*Files)(nil)
	_ store.Batcher       = (*Files)(nil)
)

// New creates one files reference store.
//
// gitRoot is the per-worktree git directory,
// and commonRoot is the common git directory.
// For repositories without a separate per-worktree git directory,
// such as bare repositories and the main worktree,
// pass the same root twice.
// Both roots must be non-nil.
//
// Labels: Deps-Borrowed, Life-Parent.
func New(gitRoot, commonRoot *os.Root, objectFormat id.ObjectFormat, options Options) *Files {
	return &Files{
		gitRoot:      gitRoot,
		commonRoot:   commonRoot,
		objectFormat: objectFormat,
		options:      options,
	}
}

// ObjectFormat returns the object format used by the store.
func (files *Files) ObjectFormat() id.ObjectFormat {
	return files.objectFormat
}

// Close releases resources held by the store itself.
//
// It does not close the roots,
// which the store does not own.
//
// Labels: MT-Unsafe.
func (files *Files) Close() error {
	return nil
}

// rootKind selects one of the store's two roots.
type rootKind uint8

const (
	rootGit rootKind = iota
	rootCommon
)

// refPath locates one loose reference file
// relative to one of the store's two roots.
type refPath struct {
	kind rootKind
	path string
}

func (files *Files) root(kind rootKind) *os.Root {
	if kind == rootCommon {
		return files.commonRoot
	}

	return files.gitRoot
}

// loosePath maps one reference name to its loose file location.
func (files *Files) loosePath(name string) refPath {
	parsed := refname.ParseWorktree(name)
	switch parsed.Type {
	case refname.WorktreeCurrent:
		return refPath{kind: rootGit, path: parsed.BareRefName}
	case refname.WorktreeMain:
		return refPath{kind: rootCommon, path: parsed.BareRefName}
	case refname.WorktreeOther:
		return refPath{
			kind: rootCommon,
			path: path.Join("worktrees", parsed.WorktreeName, parsed.BareRefName),
		}
	case refname.WorktreeShared:
		return refPath{kind: rootCommon, path: parsed.BareRefName}
	default:
		panic(fmt.Sprintf("ref/store/files: unsupported worktree ref type %d", parsed.Type))
	}
}
