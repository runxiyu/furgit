package repository

import (
	"errors"
	"fmt"
	"os"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/refstore"
	refchain "codeberg.org/lindenii/furgit/refstore/chain"
	refloose "codeberg.org/lindenii/furgit/refstore/loose"
	refpacked "codeberg.org/lindenii/furgit/refstore/packed"
	reftable "codeberg.org/lindenii/furgit/refstore/reftable"
)

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
