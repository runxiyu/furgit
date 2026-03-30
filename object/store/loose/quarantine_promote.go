package loose

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

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
		return tempRootErr
	}

	return removeErr
}

func promoteLooseQuarantine(parent *Store, tempName string, tempRoot *os.Root) error {
	entries, err := fs.ReadDir(tempRoot.FS(), ".")
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			return fmt.Errorf("objectstore/loose: quarantine contains unexpected file %q", entry.Name())
		}

		if len(entry.Name()) != 2 || !isHexString(entry.Name()) {
			return fmt.Errorf("objectstore/loose: quarantine contains invalid shard %q", entry.Name())
		}

		err := promoteLooseQuarantineShard(parent, tempName, tempRoot, entry.Name())
		if err != nil {
			return err
		}
	}

	return nil
}

func promoteLooseQuarantineShard(parent *Store, tempName string, tempRoot *os.Root, shard string) error {
	entries, err := fs.ReadDir(tempRoot.FS(), shard)
	if err != nil {
		return err
	}

	err = parent.root.MkdirAll(shard, 0o755)
	if err != nil {
		return err
	}

	wantNameLen := parent.algo.HexLen() - 2

	for _, entry := range entries {
		if entry.IsDir() {
			return fmt.Errorf("objectstore/loose: quarantine shard %q contains unexpected directory %q", shard, entry.Name())
		}

		if len(entry.Name()) != wantNameLen || !isHexString(entry.Name()) {
			return fmt.Errorf("objectstore/loose: quarantine shard %q contains invalid object path %q", shard, entry.Name())
		}

		err := promoteLooseQuarantineObject(parent.root, filepath.Join(tempName, shard, entry.Name()), filepath.Join(shard, entry.Name()))
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

	return fmt.Errorf("objectstore/loose: promote quarantine %q -> %q: %w", src, dst, err)
}

func isHexString(s string) bool {
	for _, ch := range s {
		if ('0' <= ch && ch <= '9') || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F') {
			continue
		}

		return false
	}

	return true
}
