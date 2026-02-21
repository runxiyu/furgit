// Package repository wires object and ref storage for one Git repository.
package repository

import (
	"errors"
	"fmt"
	"os"

	"codeberg.org/lindenii/furgit/config"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	objectchain "codeberg.org/lindenii/furgit/objectstore/chain"
	objectloose "codeberg.org/lindenii/furgit/objectstore/loose"
	objectpacked "codeberg.org/lindenii/furgit/objectstore/packed"
	"codeberg.org/lindenii/furgit/refstore"
	refchain "codeberg.org/lindenii/furgit/refstore/chain"
	refloose "codeberg.org/lindenii/furgit/refstore/loose"
	refpacked "codeberg.org/lindenii/furgit/refstore/packed"
	reftable "codeberg.org/lindenii/furgit/refstore/reftable"
)

// Repository is a thin composition root for repository-local stores.
//
// Open expects path to be the Git directory itself:
// a bare repository root or a non-bare ".git" directory.
type Repository struct {
	root         *os.Root
	objectsRoot  *os.Root
	packRoot     *os.Root
	reftableRoot *os.Root

	config *config.Config
	algo   objectid.Algorithm

	objects objectstore.Store
	refs    refstore.Store
}

// Open opens a repository and wires object/ref stores from its on-disk format.
func Open(path string) (repo *Repository, err error) {
	root, err := os.OpenRoot(path)
	if err != nil {
		return nil, err
	}

	repo = &Repository{root: root}
	defer func() {
		if err != nil {
			_ = repo.Close()
		}
	}()

	cfg, err := parseRepositoryConfig(root)
	if err != nil {
		return nil, err
	}
	repo.config = cfg

	algo, err := detectObjectAlgorithm(cfg)
	if err != nil {
		return nil, err
	}
	repo.algo = algo

	objects, objectsRoot, packRoot, err := openObjectStore(root, algo)
	if err != nil {
		return nil, err
	}
	repo.objects = objects
	repo.objectsRoot = objectsRoot
	repo.packRoot = packRoot

	refs, reftableRoot, err := openRefStore(root, algo)
	if err != nil {
		return nil, err
	}
	repo.refs = refs
	repo.reftableRoot = reftableRoot

	return repo, nil
}

// Algorithm returns the repository object ID algorithm.
func (repo *Repository) Algorithm() objectid.Algorithm {
	return repo.algo
}

// Config returns the parsed repository configuration snapshot.
//
// The returned pointer is owned by Repository. Callers should treat it as
// read-only.
func (repo *Repository) Config() *config.Config {
	return repo.config
}

// Objects returns the configured object store.
func (repo *Repository) Objects() objectstore.Store {
	return repo.objects
}

// Refs returns the configured ref store.
func (repo *Repository) Refs() refstore.Store {
	return repo.refs
}

// Close closes owned stores and filesystem roots.
// The behavior of the repo after Close is undefined.
func (repo *Repository) Close() error {
	var errs []error

	if repo.refs != nil {
		if err := repo.refs.Close(); err != nil {
			errs = append(errs, err)
		}
		repo.refs = nil
	}
	if repo.objects != nil {
		if err := repo.objects.Close(); err != nil {
			errs = append(errs, err)
		}
		repo.objects = nil
	}

	if repo.reftableRoot != nil {
		if err := repo.reftableRoot.Close(); err != nil {
			errs = append(errs, err)
		}
		repo.reftableRoot = nil
	}
	if repo.packRoot != nil {
		if err := repo.packRoot.Close(); err != nil {
			errs = append(errs, err)
		}
		repo.packRoot = nil
	}
	if repo.objectsRoot != nil {
		if err := repo.objectsRoot.Close(); err != nil {
			errs = append(errs, err)
		}
		repo.objectsRoot = nil
	}
	if repo.root != nil {
		if err := repo.root.Close(); err != nil {
			errs = append(errs, err)
		}
		repo.root = nil
	}

	return errors.Join(errs...)
}

func parseRepositoryConfig(root *os.Root) (*config.Config, error) {
	configFile, err := root.Open("config")
	if err != nil {
		return nil, fmt.Errorf("repository: open config: %w", err)
	}
	defer func() { _ = configFile.Close() }()

	cfg, err := config.ParseConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("repository: parse config: %w", err)
	}
	return cfg, nil
}

func detectObjectAlgorithm(cfg *config.Config) (objectid.Algorithm, error) {
	algoName := cfg.Get("extensions", "", "objectformat")
	if algoName == "" {
		algoName = objectid.AlgorithmSHA1.String()
	}
	algo, ok := objectid.ParseAlgorithm(algoName)
	if !ok {
		return objectid.AlgorithmUnknown, fmt.Errorf("repository: unsupported object format %q", algoName)
	}
	return algo, nil
}

func openObjectStore(root *os.Root, algo objectid.Algorithm) (out objectstore.Store, objectsRoot *os.Root, packRoot *os.Root, err error) {
	objectsRoot, err = root.OpenRoot("objects")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("repository: open objects: %w", err)
	}
	defer func() {
		if err != nil {
			if out != nil {
				_ = out.Close()
			}
			if packRoot != nil {
				_ = packRoot.Close()
			}
			_ = objectsRoot.Close()
		}
	}()

	looseStore, err := objectloose.New(objectsRoot, algo)
	if err != nil {
		return nil, nil, nil, err
	}
	backends := []objectstore.Store{looseStore}

	packRoot, err = objectsRoot.OpenRoot("pack")
	if err == nil {
		var packedStore *objectpacked.Store
		packedStore, err = objectpacked.New(packRoot, algo)
		if err != nil {
			return nil, nil, nil, err
		}
		backends = append(backends, packedStore)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, nil, nil, fmt.Errorf("repository: open objects/pack: %w", err)
	}
	err = nil
	out = objectchain.New(backends...)

	return out, objectsRoot, packRoot, nil
}

func openRefStore(root *os.Root, algo objectid.Algorithm) (out refstore.Store, reftableRoot *os.Root, err error) {
	var closePackedStore refstore.Store
	defer func() {
		if err != nil {
			if out != nil {
				_ = out.Close()
			}
			if closePackedStore != nil {
				_ = closePackedStore.Close()
			}
			if reftableRoot != nil {
				_ = reftableRoot.Close()
			}
		}
	}()

	hasReftable, err := hasReftableStack(root)
	if err != nil {
		return nil, nil, err
	}
	if hasReftable {
		reftableRoot, err = root.OpenRoot("reftable")
		if err != nil {
			return nil, nil, fmt.Errorf("repository: open reftable: %w", err)
		}
		var reftableStore *reftable.Store
		reftableStore, err = reftable.New(reftableRoot, algo)
		if err != nil {
			return nil, nil, err
		}
		err = nil
		out = reftableStore
		return reftableStore, reftableRoot, nil
	}

	looseStore, err := refloose.New(root, algo)
	if err != nil {
		return nil, nil, err
	}
	backends := []refstore.Store{looseStore}

	packedRefsFile, err := root.Open("packed-refs")
	if err == nil {
		packedStore, packedErr := refpacked.New(packedRefsFile, algo)
		_ = packedRefsFile.Close()
		if packedErr != nil {
			err = packedErr
			return nil, nil, err
		}
		closePackedStore = packedStore
		backends = append(backends, packedStore)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, nil, fmt.Errorf("repository: open packed-refs: %w", err)
	}
	err = nil
	out = refchain.New(backends...)
	closePackedStore = nil

	return out, nil, nil
}

func hasReftableStack(root *os.Root) (bool, error) {
	_, err := root.Stat("reftable/tables.list")
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("repository: stat reftable/tables.list: %w", err)
}
