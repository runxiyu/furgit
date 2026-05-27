package loose

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"os"

	objectstore "lindenii.org/go/furgit/object/store"
)

// BeginObjectQuarantine creates one quarantined loose store rooted privately
// beneath the destination loose root.
//
// Labels: Deps-Borrowed, Life-Parent, Close-No.
func (store *Store) BeginObjectQuarantine(_ objectstore.ObjectQuarantineOptions) (objectstore.ObjectQuarantine, error) {
	tempName, tempRoot, err := createLooseQuarantineRoot(store.root)
	if err != nil {
		return nil, err
	}

	quarantineStore, err := New(tempRoot, store.algo)
	if err != nil {
		_ = tempRoot.Close()
		_ = store.root.RemoveAll(tempName)

		return nil, err
	}

	return &objectQuarantine{
		Store:    quarantineStore,
		parent:   store,
		tempName: tempName,
		tempRoot: tempRoot,
	}, nil
}

func createLooseQuarantineRoot(parent *os.Root) (string, *os.Root, error) {
	for range 32 {
		name := "tmp_looseq_" + rand.Text()

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

	return "", nil, fmt.Errorf("objectstore/loose: unable to create quarantine directory")
}
