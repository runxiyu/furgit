package packed

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"os"

	objectstore "lindenii.org/go/furgit/object/store"
)

// BeginPackQuarantine creates one quarantined packed store rooted privately
// beneath the destination pack root.
//
// Labels: Deps-Borrowed, Life-Parent, Close-No.
func (store *Store) BeginPackQuarantine(_ objectstore.PackQuarantineOptions) (objectstore.PackQuarantine, error) {
	tempName, tempRoot, err := createPackQuarantineRoot(store.root)
	if err != nil {
		return nil, err
	}

	quarantineStore, err := New(tempRoot, store.algo, store.opts)
	if err != nil {
		_ = tempRoot.Close()
		_ = store.root.RemoveAll(tempName)

		return nil, err
	}

	return &packQuarantine{
		Store:    quarantineStore,
		parent:   store,
		tempName: tempName,
		tempRoot: tempRoot,
	}, nil
}

func createPackQuarantineRoot(parent *os.Root) (string, *os.Root, error) {
	for range 32 {
		name := "tmp_packq_" + rand.Text()

		err := parent.Mkdir(name, 0o700)
		if err == nil {
			root, err := parent.OpenRoot(name)
			if err == nil {
				return name, root, nil
			}

			_ = parent.RemoveAll(name)

			return "", nil, err
		}

		if errors.Is(err, fs.ErrExist) {
			continue
		}

		return "", nil, err
	}

	return "", nil, fmt.Errorf("packed: unable to create quarantine directory")
}
