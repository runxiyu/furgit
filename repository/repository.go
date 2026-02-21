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
	config *config.Config
	algo   objectid.Algorithm

	objects objectstore.Store
	refs    refstore.Store
}

// Open opens a repository and wires object/ref stores from its on-disk format.
func Open(path string) (repo *Repository, err error) {
	setupRoot, err := os.OpenRoot(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = setupRoot.Close() }()

	repo = &Repository{}
	defer func() {
		if err != nil {
			_ = repo.Close()
		}
	}()

	cfg, err := parseRepositoryConfig(setupRoot)
	if err != nil {
		return nil, err
	}
	repo.config = cfg

	algo, err := detectObjectAlgorithm(cfg)
	if err != nil {
		return nil, err
	}
	repo.algo = algo

	objects, err := openObjectStore(path, algo)
	if err != nil {
		return nil, err
	}
	repo.objects = objects

	refs, err := openRefStore(path, algo)
	if err != nil {
		return nil, err
	}
	repo.refs = refs

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
	}
	if repo.objects != nil {
		if err := repo.objects.Close(); err != nil {
			errs = append(errs, err)
		}
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

func openObjectStore(path string, algo objectid.Algorithm) (out objectstore.Store, err error) {
	repoRoot, err := os.OpenRoot(path)
	if err != nil {
		return nil, fmt.Errorf("repository: open root: %w", err)
	}
	defer func() { _ = repoRoot.Close() }()

	objectsRoot, err := repoRoot.OpenRoot("objects")
	if err != nil {
		return nil, fmt.Errorf("repository: open objects: %w", err)
	}
	var packRoot *os.Root
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
		return nil, err
	}
	backends := []objectstore.Store{looseStore}

	packRoot, err = objectsRoot.OpenRoot("pack")
	if err == nil {
		var packedStore *objectpacked.Store
		packedStore, err = objectpacked.New(packRoot, algo)
		if err != nil {
			return nil, err
		}
		backends = append(backends, packedStore)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("repository: open objects/pack: %w", err)
	}
	err = nil
	out = objectchain.New(backends...)

	return out, nil
}

func openRefStore(path string, algo objectid.Algorithm) (out refstore.Store, err error) {
	metaRoot, err := os.OpenRoot(path)
	if err != nil {
		return nil, fmt.Errorf("repository: open root: %w", err)
	}
	defer func() { _ = metaRoot.Close() }()

	hasReftable, err := hasReftableStack(metaRoot)
	if err != nil {
		return nil, err
	}
	if hasReftable {
		reftableRoot, err := metaRoot.OpenRoot("reftable")
		if err != nil {
			return nil, fmt.Errorf("repository: open reftable: %w", err)
		}
		reftableStore, err := reftable.New(reftableRoot, algo)
		if err != nil {
			_ = reftableRoot.Close()
			return nil, err
		}
		return reftableStore, nil
	}

	looseRoot, err := os.OpenRoot(path)
	if err != nil {
		return nil, fmt.Errorf("repository: open root for loose refs: %w", err)
	}
	looseStore, err := refloose.New(looseRoot, algo)
	if err != nil {
		_ = looseRoot.Close()
		return nil, err
	}
	backends := []refstore.Store{looseStore}

	packedRefsFile, err := metaRoot.Open("packed-refs")
	if err == nil {
		packedStore, packedErr := refpacked.New(packedRefsFile, algo)
		_ = packedRefsFile.Close()
		if packedErr != nil {
			_ = looseStore.Close()
			return nil, packedErr
		}
		backends = append(backends, packedStore)
	} else if !errors.Is(err, os.ErrNotExist) {
		_ = looseStore.Close()
		return nil, fmt.Errorf("repository: open packed-refs: %w", err)
	}

	return refchain.New(backends...), nil
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
