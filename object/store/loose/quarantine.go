package loose

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"lindenii.org/go/furgit/object/store"
)

var (
	_ store.ObjectQuarantiner = (*Loose)(nil)
	_ store.ObjectQuarantine  = (*objectQuarantine)(nil)
)

// objectQuarantine is one quarantined loose store
// rooted privately beneath a destination loose root.
type objectQuarantine struct {
	*Loose

	parent   *Loose
	tempName string
	tempRoot *os.Root
}

// BeginObjectQuarantine creates one quarantined loose store rooted privately
// beneath the destination loose root.
//
// Labels: Deps-Borrowed, Life-Parent, Close-No.
func (loose *Loose) BeginObjectQuarantine(_ store.ObjectQuarantineOptions) (store.ObjectQuarantine, error) {
	tempName, tempRoot, err := createLooseQuarantineRoot(loose.root)
	if err != nil {
		return nil, err
	}

	quarantineStore, err := New(tempRoot, loose.objectFormat)
	if err != nil {
		_ = tempRoot.Close()
		_ = loose.root.RemoveAll(tempName)

		return nil, err
	}

	return &objectQuarantine{
		Loose:    quarantineStore,
		parent:   loose,
		tempName: tempName,
		tempRoot: tempRoot,
	}, nil
}

// Discard removes the quarantine and invalidates the receiver.
func (quarantine *objectQuarantine) Discard() error {
	closeErr := quarantine.Close()
	tempRootErr := quarantine.tempRoot.Close()
	removeErr := quarantine.parent.root.RemoveAll(quarantine.tempName)

	if closeErr != nil {
		return closeErr
	}

	if tempRootErr != nil {
		return fmt.Errorf("object/store/loose: %w", tempRootErr)
	}

	if removeErr != nil {
		return fmt.Errorf("object/store/loose: %w", removeErr)
	}

	return nil
}

// Promote publishes all quarantined loose objects into the parent loose store
// and invalidates the receiver.
func (quarantine *objectQuarantine) Promote() error {
	closeErr := quarantine.Close()
	promoteErr := promoteLooseQuarantine(quarantine.parent, quarantine.tempName, quarantine.tempRoot)
	tempRootErr := quarantine.tempRoot.Close()
	removeErr := quarantine.parent.root.RemoveAll(quarantine.tempName)

	if closeErr != nil {
		return closeErr
	}

	if promoteErr != nil {
		return promoteErr
	}

	if tempRootErr != nil {
		return fmt.Errorf("object/store/loose: %w", tempRootErr)
	}

	if removeErr != nil {
		return fmt.Errorf("object/store/loose: %w", removeErr)
	}

	return nil
}

func createLooseQuarantineRoot(parent *os.Root) (string, *os.Root, error) {
	var lastErr error

	for range 32 {
		name := "tmp_looseq_" + rand.Text()

		err := parent.Mkdir(name, 0o700)
		if err == nil {
			root, err := parent.OpenRoot(name)
			if err == nil {
				return name, root, nil
			}

			_ = parent.RemoveAll(name)

			return "", nil, fmt.Errorf("object/store/loose: %w", err)
		}

		lastErr = err

		if errors.Is(err, fs.ErrExist) {
			continue
		}

		return "", nil, fmt.Errorf("object/store/loose: %w", err)
	}

	return "", nil, fmt.Errorf("object/store/loose: failed to create quarantine directory: %w", lastErr)
}

func promoteLooseQuarantine(parent *Loose, tempName string, tempRoot *os.Root) error {
	entries, err := fs.ReadDir(tempRoot.FS(), ".")
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("object/store/loose: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			return fmt.Errorf("%w: quarantine contains unexpected file %q", store.ErrInvalidObject, entry.Name())
		}

		err := promoteLooseQuarantineShard(parent, tempName, tempRoot, entry.Name())
		if err != nil {
			return err
		}
	}

	return nil
}

func promoteLooseQuarantineShard(parent *Loose, tempName string, tempRoot *os.Root, shard string) error {
	entries, err := fs.ReadDir(tempRoot.FS(), shard)
	if err != nil {
		return fmt.Errorf("object/store/loose: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			return fmt.Errorf("%w: quarantine shard %q contains unexpected directory %q", store.ErrInvalidObject, shard, entry.Name())
		}

		objectID, err := parent.objectFormat.FromString(shard + entry.Name())
		if err != nil {
			return fmt.Errorf("%w: quarantine shard %q contains invalid object %q: %w", store.ErrInvalidObject, shard, entry.Name(), err)
		}

		dst, err := parent.objectPath(objectID)
		if err != nil {
			return err
		}

		err = parent.root.MkdirAll(shard, 0o755)
		if err != nil {
			return fmt.Errorf("object/store/loose: %w", err)
		}

		err = promoteLooseQuarantineObject(parent.root, filepath.Join(tempName, shard, entry.Name()), dst)
		if err != nil {
			return err
		}
	}

	return nil
}

func promoteLooseQuarantineObject(root *os.Root, src, dst string) error {
	err := root.Link(src, dst)
	if err == nil {
		_ = root.Remove(src)

		return nil
	}

	if errors.Is(err, fs.ErrExist) {
		_ = root.Remove(src)

		return nil
	}

	return fmt.Errorf("object/store/loose: promote quarantine %q -> %q: %w", src, dst, err)
}
