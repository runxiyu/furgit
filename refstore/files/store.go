// Package files provides one Git files ref store with loose-over-packed reads
// and transaction-coordinated updates.
package files

import (
	"errors"
	"io"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"codeberg.org/lindenii/furgit/config"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/ref/refname"
	"codeberg.org/lindenii/furgit/refstore"
)

// Store reads and writes one Git files ref namespace rooted at one repository
// gitdir plus its commondir.
//
// Store owns both roots and closes them in Close.
type Store struct {
	gitRoot    *os.Root
	commonRoot *os.Root
	algo       objectid.Algorithm
	lockRand   *rand.Rand

	packedRefsTimeout time.Duration
}

var (
	_ refstore.ReadingStore       = (*Store)(nil)
	_ refstore.TransactionalStore = (*Store)(nil)
)

type rootKind uint8

const (
	rootGit rootKind = iota
	rootCommon
)

type refPath struct {
	root rootKind
	path string
}

// New creates one files ref store rooted at one repository gitdir.
func New(root *os.Root, algo objectid.Algorithm) (*Store, error) {
	if algo.Size() == 0 {
		return nil, objectid.ErrInvalidAlgorithm
	}

	commonRoot, err := openCommonRoot(root)
	if err != nil {
		return nil, err
	}

	return &Store{
		gitRoot:           root,
		commonRoot:        commonRoot,
		algo:              algo,
		lockRand:          rand.New(rand.NewSource(time.Now().UnixNano())), //nolint:gosec
		packedRefsTimeout: detectPackedRefsTimeout(commonRoot),
	}, nil
}

// Close releases resources associated with the store.
func (store *Store) Close() error {
	err := store.gitRoot.Close()
	commonErr := store.commonRoot.Close()

	if err != nil {
		return err
	}

	return commonErr
}

func openCommonRoot(gitRoot *os.Root) (*os.Root, error) {
	content, err := gitRoot.ReadFile("commondir")
	if err != nil {
		if errorsIsNotExist(err) {
			return gitRoot.OpenRoot(".")
		}

		return nil, err
	}

	commonDir := strings.TrimSpace(string(content))
	if commonDir == "" {
		return nil, os.ErrNotExist
	}

	if filepath.IsAbs(commonDir) {
		return os.OpenRoot(commonDir)
	}

	// This is okay because that's how Git defines it anyway.
	return os.OpenRoot(filepath.Join(gitRoot.Name(), commonDir))
}

func (store *Store) rootFor(kind rootKind) *os.Root {
	if kind == rootCommon {
		return store.commonRoot
	}

	return store.gitRoot
}

func (store *Store) loosePath(name string) refPath {
	parsed := refname.ParseWorktree(name)
	switch parsed.Type {
	case refname.WorktreeCurrent:
		return refPath{root: rootGit, path: parsed.BareRefName}
	case refname.WorktreeMain, refname.WorktreeShared:
		return refPath{root: rootCommon, path: parsed.BareRefName}
	case refname.WorktreeOther:
		return refPath{
			root: rootCommon,
			path: path.Join("worktrees", parsed.WorktreeName, parsed.BareRefName),
		}
	default:
		return refPath{root: rootCommon, path: name}
	}
}

func detectPackedRefsTimeout(commonRoot *os.Root) time.Duration {
	const defaultTimeout = time.Second

	file, err := commonRoot.Open("config")
	if err != nil {
		return defaultTimeout
	}

	defer func() { _ = file.Close() }()

	cfg, err := config.ParseConfig(file)
	if err != nil && !errors.Is(err, io.EOF) {
		return defaultTimeout
	}

	timeoutValue, err := cfg.Lookup("core", "", "packedrefstimeout").Int()
	if err != nil {
		return defaultTimeout
	}

	return time.Duration(timeoutValue) * time.Millisecond
}
