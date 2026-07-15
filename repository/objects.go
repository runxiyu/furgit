package repository

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/store/dual"
	"lindenii.org/go/furgit/object/store/loose"
	"lindenii.org/go/furgit/object/store/packed"
)

type objectStores struct {
	dual     *dual.Dual
	root     *os.Root
	packRoot *os.Root
	loose    *loose.Loose
	packed   *packed.Packed
}

// Objects returns the object store of the repository.
//
// Callers who want typed object values
// should usually prefer [Repository.Fetcher].
//
// Labels: Life-Parent.
func (repo *Repository) Objects() interface {
	store.ObjectReader
	store.ObjectWriter
	store.PackWriter
	store.ObjectQuarantiner
	store.PackQuarantiner
	store.CoordinatedQuarantiner
} {
	return repo.objects
}

func openObjects(commonRoot *os.Root, objectFormat id.ObjectFormat) (*objectStores, error) {
	root, err := commonRoot.OpenRoot("objects")
	if err != nil {
		return nil, fmt.Errorf("repository: open objects: %w", err)
	}

	looseStore, err := loose.New(root, objectFormat)
	if err != nil {
		_ = root.Close()

		return nil, fmt.Errorf("repository: open loose objects: %w", err)
	}

	err = root.Mkdir("pack", 0o755)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		_ = looseStore.Close()
		_ = root.Close()

		return nil, fmt.Errorf("repository: create objects/pack: %w", err)
	}

	packRoot, err := root.OpenRoot("pack")
	if err != nil {
		_ = looseStore.Close()
		_ = root.Close()

		return nil, fmt.Errorf("repository: open objects/pack: %w", err)
	}

	packedStore, err := packed.New(packRoot, objectFormat)
	if err != nil {
		_ = packRoot.Close()
		_ = looseStore.Close()
		_ = root.Close()

		return nil, fmt.Errorf("repository: open packed objects: %w", err)
	}

	return &objectStores{
		dual:     dual.New(looseStore, packedStore),
		root:     root,
		packRoot: packRoot,
		loose:    looseStore,
		packed:   packedStore,
	}, nil
}
