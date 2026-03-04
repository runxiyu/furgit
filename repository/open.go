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

// Open opens a repository and wires object/ref stores from its on-disk format.
//
// Open borrows root during construction and does not close it.
func Open(root *os.Root) (repo *Repository, err error) {
	repo = &Repository{}

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

	objects, objectsLooseForWritingOnly, err := openObjectStore(root, algo)
	if err != nil {
		return nil, err
	}

	repo.objects = objects
	repo.objectsLooseForWritingOnly = objectsLooseForWritingOnly

	refs, err := openRefStore(root, algo)
	if err != nil {
		return nil, err
	}

	repo.refs = refs

	return repo, nil
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
	algoName := cfg.Lookup("extensions", "", "objectformat").Value
	if algoName == "" {
		algoName = objectid.AlgorithmSHA1.String()
	}

	algo, ok := objectid.ParseAlgorithm(algoName)
	if !ok {
		return objectid.AlgorithmUnknown, fmt.Errorf("repository: unsupported object format %q", algoName)
	}

	return algo, nil
}

func openObjectStore(root *os.Root, algo objectid.Algorithm) (objectstore.Store, *objectloose.Store, error) {
	objectsRoot, err := root.OpenRoot("objects")
	if err != nil {
		return nil, nil, fmt.Errorf("repository: open objects: %w", err)
	}

	looseStore, err := objectloose.New(objectsRoot, algo)
	if err != nil {
		return nil, nil, err
	}

	backends := []objectstore.Store{looseStore}

	packRoot, err := objectsRoot.OpenRoot("pack")
	if err == nil {
		var packedStore *objectpacked.Store

		packedStore, err = objectpacked.New(packRoot, algo)
		if err != nil {
			_ = looseStore.Close()

			return nil, nil, err
		}

		backends = append(backends, packedStore)
	} else if !errors.Is(err, os.ErrNotExist) {
		_ = looseStore.Close()

		return nil, nil, fmt.Errorf("repository: open objects/pack: %w", err)
	}

	objectsChain := objectchain.New(backends...)

	objectsRootForWriting, err := root.OpenRoot("objects")
	if err != nil {
		_ = objectsChain.Close()

		return nil, nil, fmt.Errorf("repository: open objects for loose writing: %w", err)
	}

	objectsLooseForWritingOnly, err := objectloose.New(objectsRootForWriting, algo)
	if err != nil {
		_ = objectsRootForWriting.Close()
		_ = objectsChain.Close()

		return nil, nil, err
	}

	return objectsChain, objectsLooseForWritingOnly, nil
}

func openRefStore(root *os.Root, algo objectid.Algorithm) (out refstore.Store, err error) {
	hasReftable, err := hasReftableStack(root)
	if err != nil {
		return nil, err
	}

	if hasReftable {
		reftableRoot, err := root.OpenRoot("reftable")
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

	looseRoot, err := root.OpenRoot(".")
	if err != nil {
		return nil, fmt.Errorf("repository: open root for loose refs: %w", err)
	}

	looseStore, err := refloose.New(looseRoot, algo)
	if err != nil {
		_ = looseRoot.Close()

		return nil, err
	}

	backends := []refstore.Store{looseStore}

	_, err = root.Stat("packed-refs")
	if err == nil {
		packedStore, packedErr := refpacked.New(root, algo)
		if packedErr != nil {
			_ = looseStore.Close()

			return nil, packedErr
		}

		backends = append(backends, packedStore)
	} else if !errors.Is(err, os.ErrNotExist) {
		_ = looseStore.Close()

		return nil, fmt.Errorf("repository: stat packed-refs: %w", err)
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
