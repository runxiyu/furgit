package repository

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"slices"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/store/chain"
	"lindenii.org/go/furgit/object/store/dual"
	"lindenii.org/go/furgit/object/store/loose"
	"lindenii.org/go/furgit/object/store/packed"
	"lindenii.org/go/furgit/object/typ"
)

type objectStores struct {
	store      *objectStore
	root       *os.Root
	packRoot   *os.Root
	loose      *loose.Loose
	packed     *packed.Packed
	alternates []*alternate
}

// objectStore reads objects from the repository and its alternates,
// while confining writes to the repository itself.
type objectStore struct {
	*dual.Dual

	// reader covers the repository and its alternates,
	// and so replaces every read the embedded store would serve alone.
	reader *chain.Chain
}

var (
	_ store.ObjectReader           = (*objectStore)(nil)
	_ store.ObjectWriter           = (*objectStore)(nil)
	_ store.PackWriter             = (*objectStore)(nil)
	_ store.ObjectQuarantiner      = (*objectStore)(nil)
	_ store.PackQuarantiner        = (*objectStore)(nil)
	_ store.CoordinatedQuarantiner = (*objectStore)(nil)
)

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

func (objects *objectStore) ReadBytesFull(objectID id.ObjectID) ([]byte, error) {
	return objects.reader.ReadBytesFull(objectID) //nolint:wrapcheck
}

func (objects *objectStore) ReadBytesContent(objectID id.ObjectID) (typ.Type, []byte, error) {
	return objects.reader.ReadBytesContent(objectID) //nolint:wrapcheck
}

func (objects *objectStore) ReadReaderFull(objectID id.ObjectID) (io.ReadCloser, error) {
	return objects.reader.ReadReaderFull(objectID) //nolint:wrapcheck
}

func (objects *objectStore) ReadReaderContent(objectID id.ObjectID) (typ.Type, int, io.ReadCloser, error) {
	return objects.reader.ReadReaderContent(objectID) //nolint:wrapcheck
}

func (objects *objectStore) ReadSize(objectID id.ObjectID) (int, error) {
	return objects.reader.ReadSize(objectID) //nolint:wrapcheck
}

func (objects *objectStore) ReadHeader(objectID id.ObjectID) (typ.Type, int, error) {
	return objects.reader.ReadHeader(objectID) //nolint:wrapcheck
}

func (objects *objectStore) Refresh() error {
	return objects.reader.Refresh() //nolint:wrapcheck
}

func openObjects(
	commonRoot *os.Root,
	objectFormat id.ObjectFormat,
	options Options,
) (stores *objectStores, err error) {
	closers := []func() error{}

	defer func() {
		if err == nil {
			return
		}

		for _, closer := range slices.Backward(closers) {
			_ = closer()
		}
	}()

	root, err := commonRoot.OpenRoot("objects")
	if err != nil {
		return nil, fmt.Errorf("repository: open objects: %w", err)
	}

	closers = append(closers, root.Close)

	looseStore, err := loose.New(root, objectFormat)
	if err != nil {
		return nil, fmt.Errorf("repository: open loose objects: %w", err)
	}

	closers = append(closers, looseStore.Close)

	err = root.Mkdir("pack", 0o755)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return nil, fmt.Errorf("repository: create objects/pack: %w", err)
	}

	packRoot, err := root.OpenRoot("pack")
	if err != nil {
		return nil, fmt.Errorf("repository: open objects/pack: %w", err)
	}

	closers = append(closers, packRoot.Close)

	packedStore, err := packed.New(packRoot, objectFormat)
	if err != nil {
		return nil, fmt.Errorf("repository: open packed objects: %w", err)
	}

	closers = append(closers, packedStore.Close)

	local := dual.New(looseStore, packedStore)

	alternates, err := openAlternates(root, objectFormat, options)
	if err != nil {
		return nil, err
	}

	for _, alt := range alternates {
		closers = append(closers, alt.close)
	}

	readers := []store.ObjectReader{local}
	for _, alt := range alternates {
		readers = append(readers, alt.readers()...)
	}

	return &objectStores{
		store:      &objectStore{Dual: local, reader: chain.New(readers...)},
		root:       root,
		packRoot:   packRoot,
		loose:      looseStore,
		packed:     packedStore,
		alternates: alternates,
	}, nil
}
