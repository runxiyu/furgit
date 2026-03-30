package packed

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"
)

// Promote publishes all finalized pack artifacts in the quarantine into the
// parent packed store and invalidates the receiver.
func (quarantine *packQuarantine) Promote() error {
	closeErr := quarantine.Close()
	promoteErr := promotePackQuarantine(quarantine.parent.root, quarantine.tempName, quarantine.tempRoot)
	tempRootErr := quarantine.tempRoot.Close()
	removeErr := quarantine.parent.root.RemoveAll(quarantine.tempName)

	if closeErr != nil {
		return closeErr
	}

	if tempRootErr != nil {
		return tempRootErr
	}

	if promoteErr != nil {
		return promoteErr
	}

	return removeErr
}

func promotePackQuarantine(parent *os.Root, tempName string, tempRoot *os.Root) error {
	entries, err := fs.ReadDir(tempRoot.FS(), ".")
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	slices.SortFunc(entries, func(left, right fs.DirEntry) int {
		return packPromotionPriority(left.Name()) - packPromotionPriority(right.Name())
	})

	for _, entry := range entries {
		if entry.IsDir() {
			return fmt.Errorf("packed: quarantine contains unexpected directory %q", entry.Name())
		}

		err := promotePackQuarantineFile(parent, tempName, entry.Name())
		if err != nil {
			return err
		}
	}

	return nil
}

func promotePackQuarantineFile(parent *os.Root, tempName, name string) error {
	src := tempName + "/" + name

	err := parent.Link(src, name)
	if err == nil {
		_ = parent.Remove(src)

		return nil
	}

	if errors.Is(err, fs.ErrExist) {
		_ = parent.Remove(src)

		return nil
	}

	return fmt.Errorf("packed: promote quarantine %q -> %q: %w", src, name, err)
}

func packPromotionPriority(name string) int {
	switch {
	case strings.HasPrefix(name, "pack-") && strings.HasSuffix(name, ".pack"):
		return 1
	case strings.HasPrefix(name, "pack-") && strings.HasSuffix(name, ".rev"):
		return 2
	case strings.HasPrefix(name, "pack-") && strings.HasSuffix(name, ".idx"):
		return 3
	default:
		return 0
	}
}
