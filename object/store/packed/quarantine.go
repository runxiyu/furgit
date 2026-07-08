package packed

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"

	"lindenii.org/go/furgit/object/store"
)

var (
	_ store.PackQuarantiner = (*Packed)(nil)
	_ store.PackQuarantine  = (*packQuarantine)(nil)
)

var errQuarantineNamesExhausted = errors.New("object/store/packed: exhausted quarantine directory names")

// packQuarantine is one quarantined packed store
// rooted privately beneath a destination pack root.
type packQuarantine struct {
	*Packed

	parent *Packed

	tempName string
	tempRoot *os.Root
}

// BeginPackQuarantine creates one quarantined packed store
// rooted privately beneath the destination pack root.
//
// Labels: Deps-Borrowed, Life-Parent.
func (packed *Packed) BeginPackQuarantine(_ store.PackQuarantineOptions) (store.PackQuarantine, error) { //nolint:ireturn
	tempName, tempRoot, err := createPackQuarantineRoot(packed.root)
	if err != nil {
		return nil, err
	}

	quarantineStore, err := New(tempRoot, packed.objectFormat)
	if err != nil {
		_ = tempRoot.Close()
		_ = packed.root.RemoveAll(tempName)

		return nil, err
	}

	return &packQuarantine{
		Packed:   quarantineStore,
		parent:   packed,
		tempName: tempName,
		tempRoot: tempRoot,
	}, nil
}

// Promote publishes the quarantined pack artifacts into the parent store,
// refreshes the parent so the objects become available,
// and invalidates the receiver.
func (quarantine *packQuarantine) Promote() error {
	closeErr := quarantine.Close()
	promoteErr := quarantine.promoteAll()

	var refreshErr error
	if promoteErr == nil {
		refreshErr = quarantine.parent.Refresh()
	}

	tempRootErr := quarantine.tempRoot.Close()
	removeErr := quarantine.parent.root.RemoveAll(quarantine.tempName)

	return errors.Join(closeErr, promoteErr, refreshErr, tempRootErr, removeErr)
}

// Discard removes the quarantine and invalidates the receiver.
func (quarantine *packQuarantine) Discard() error {
	closeErr := quarantine.Close()
	tempRootErr := quarantine.tempRoot.Close()
	removeErr := quarantine.parent.root.RemoveAll(quarantine.tempName)

	return errors.Join(closeErr, tempRootErr, removeErr)
}

// promoteAll links every pack artifact in the quarantine into the parent store,
// in pack/rev/idx dependency order.
func (quarantine *packQuarantine) promoteAll() error {
	entries, err := fs.ReadDir(quarantine.tempRoot.FS(), ".")
	if err != nil {
		return fmt.Errorf("object/store/packed: %w", err)
	}

	slices.SortFunc(entries, func(left, right fs.DirEntry) int {
		return packPromotionPriority(left.Name()) - packPromotionPriority(right.Name())
	})

	var created []string

	for _, entry := range entries {
		linked, err := quarantine.promoteFile(entry.Name())
		if err != nil {
			for _, v := range slices.Backward(created) {
				_ = quarantine.parent.root.Remove(v)
			}

			return err
		}

		if linked {
			created = append(created, entry.Name())
		}
	}

	return nil
}

// promoteFile links one quarantined artifact into the parent store
// and reports whether the destination was newly created.
// A pre-existing destination is treated as success; rollback must not remove it.
func (quarantine *packQuarantine) promoteFile(name string) (bool, error) {
	src := quarantine.tempName + "/" + name

	err := quarantine.parent.root.Link(src, name)

	switch {
	case err == nil:
		_ = quarantine.parent.root.Remove(src)

		return true, nil
	case errors.Is(err, fs.ErrExist):
		_ = quarantine.parent.root.Remove(src)

		return false, nil
	default:
		return false, fmt.Errorf("object/store/packed: promoting %q: %w", name, err)
	}
}

// createPackQuarantineRoot creates a private quarantine directory beneath parent
// and returns its name and an os.Root over it.
func createPackQuarantineRoot(parent *os.Root) (string, *os.Root, error) {
	for range 32 {
		name := "tmp_packq_" + rand.Text()

		err := parent.Mkdir(name, 0o700)
		if err != nil {
			if errors.Is(err, fs.ErrExist) {
				continue
			}

			return "", nil, fmt.Errorf("object/store/packed: %w", err)
		}

		root, err := parent.OpenRoot(name)
		if err != nil {
			_ = parent.RemoveAll(name)

			return "", nil, fmt.Errorf("object/store/packed: %w", err)
		}

		return name, root, nil
	}

	return "", nil, errQuarantineNamesExhausted
}

// packPromotionPriority orders pack artifacts
// so that data files are linked before the index that publishes them.
func packPromotionPriority(name string) int {
	switch {
	case strings.HasPrefix(name, "pack-") && strings.HasSuffix(name, ".pack"):
		return 1
	case strings.HasPrefix(name, "pack-") && strings.HasSuffix(name, ".rev"):
		return 2
	case strings.HasPrefix(name, "pack-") && strings.HasSuffix(name, ".bloom"):
		return 2
	case strings.HasPrefix(name, "pack-") && strings.HasSuffix(name, ".idx"):
		return 3
	default:
		return 0
	}
}
