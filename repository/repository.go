package repository

import (
	"errors"
	"os"

	"lindenii.org/go/furgit/config"
	"lindenii.org/go/furgit/object/fetch"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store/dual"
	"lindenii.org/go/furgit/object/store/loose"
	"lindenii.org/go/furgit/object/store/packed"
	"lindenii.org/go/furgit/ref/store/files"
)

// Repository composes the stores and helpers
// of one on-disk Git repository.
//
// Labels: MT-Safe, Deps-Borrowed, Life-Parent, Close-Caller.
type Repository struct {
	config       *config.Config
	objectFormat id.ObjectFormat

	objects *dual.Dual
	fetcher *fetch.Fetcher

	// objectsRoot and objectsPackRoot are opened by Open,
	// so unlike the roots passed to Open,
	// they are owned, and closed by Close.
	objectsRoot     *os.Root
	objectsPackRoot *os.Root

	objectsLoose  *loose.Loose
	objectsPacked *packed.Packed
	refs          *files.Files
}

// Open opens the repository whose Git directory is gitRoot
// and whose common directory is commonRoot.
//
// gitRoot is the per-worktree Git directory,
// and commonRoot is the common Git directory.
//
// For repositories without a separate per-worktree Git directory,
// such as bare repositories and the main worktree,
// pass the same root twice.
// For a linked worktree,
// gitRoot is its ".git/worktrees/<name>" directory
// and commonRoot is the ".git" directory it was created from.
// Both roots must be non-nil.
//
// Labels: Deps-Borrowed, Life-Parent, Close-Caller.
func Open(gitRoot, commonRoot *os.Root) (*Repository, error) {
	common, err := parseConfig(commonRoot)
	if err != nil {
		return nil, err
	}

	objectFormat, err := detectObjectFormat(common)
	if err != nil {
		return nil, err
	}

	cfg, err := effectiveConfig(gitRoot, common)
	if err != nil {
		return nil, err
	}

	refOptions, err := detectRefOptions(cfg)
	if err != nil {
		return nil, err
	}

	objects, err := openObjects(commonRoot, objectFormat)
	if err != nil {
		return nil, err
	}

	return &Repository{
		config:          cfg,
		objectFormat:    objectFormat,
		objects:         objects.dual,
		fetcher:         fetch.New(objects.dual),
		objectsRoot:     objects.root,
		objectsPackRoot: objects.packRoot,
		objectsLoose:    objects.loose,
		objectsPacked:   objects.packed,
		refs:            files.New(gitRoot, commonRoot, objectFormat, refOptions),
	}, nil
}

// Close closes the stores and roots that the repository owns.
//
// It does not close the roots passed to [Open].
//
// Labels: MT-Unsafe, Idem-2UB.
func (repo *Repository) Close() error {
	return errors.Join(
		repo.objectsPacked.Close(),
		repo.objectsLoose.Close(),
		repo.refs.Close(),
		repo.objectsPackRoot.Close(),
		repo.objectsRoot.Close(),
	)
}
