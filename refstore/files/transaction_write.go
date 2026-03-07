package files

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

func (tx *Transaction) verifyCurrent(item preparedTxOp) error {
	switch item.op.kind {
	case txCreate:
		if item.target.ref.kind != directMissing {
			return fmt.Errorf("refstore/files: reference %q already exists", item.target.name)
		}

		return nil
	case txUpdate, txDelete, txVerify:
		if item.target.ref.kind == directMissing {
			return fmt.Errorf("refstore/files: reference %q is missing", item.target.name)
		}

		if item.target.ref.kind != directDetached {
			return fmt.Errorf("refstore/files: reference %q is not detached", item.target.name)
		}

		if item.target.ref.id != item.op.oldID {
			return fmt.Errorf("refstore/files: reference %q is at %s but expected %s", item.target.name, item.target.ref.id, item.op.oldID)
		}

		return nil
	case txCreateSymbolic:
		if item.target.ref.kind != directMissing {
			return fmt.Errorf("refstore/files: reference %q already exists", item.target.name)
		}

		return nil
	case txUpdateSymbolic, txDeleteSymbolic, txVerifySymbolic:
		if item.target.ref.kind == directMissing {
			return fmt.Errorf("refstore/files: symbolic reference %q is missing", item.target.name)
		}

		if item.target.ref.kind != directSymbolic {
			return fmt.Errorf("refstore/files: reference %q is not symbolic", item.target.name)
		}

		if strings.TrimSpace(item.target.ref.target) != strings.TrimSpace(item.op.oldTarget) {
			return fmt.Errorf("refstore/files: reference %q points at %q, expected %q", item.target.name, item.target.ref.target, item.op.oldTarget)
		}

		return nil
	default:
		return fmt.Errorf("refstore/files: unsupported transaction operation %d", item.op.kind)
	}
}

func (tx *Transaction) writeLoose(item preparedTxOp) error {
	root := tx.store.rootFor(item.target.loc.root)
	lockName := item.target.loc.path + ".lock"

	lock, err := root.OpenFile(lockName, os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	var content string

	switch item.op.kind {
	case txCreate, txUpdate:
		content = item.op.newID.String() + "\n"
	case txCreateSymbolic, txUpdateSymbolic:
		content = "ref: " + strings.TrimSpace(item.op.newTarget) + "\n"
	case txDelete, txVerify, txDeleteSymbolic, txVerifySymbolic:
	default:
		_ = lock.Close()

		return fmt.Errorf("refstore/files: unsupported write operation %d", item.op.kind)
	}

	_, err = lock.WriteString(content)
	if err != nil {
		_ = lock.Close()

		return err
	}

	err = lock.Close()
	if err != nil {
		return err
	}

	dir := path.Dir(item.target.loc.path)
	if dir != "." {
		err = root.MkdirAll(dir, 0o755)
		if err != nil {
			return err
		}
	}

	err = tx.removeEmptyDirTree(item.target.loc)
	if err != nil {
		return err
	}

	return root.Rename(lockName, item.target.loc.path)
}

func (tx *Transaction) applyPackedDeletes(prepared []preparedTxOp) error {
	_, err := tx.store.commonRoot.Stat("packed-refs.lock")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	packed, err := tx.store.readPackedRefs()
	if err != nil {
		return err
	}

	deleted := make(map[string]struct{})
	needed := false

	for _, item := range prepared {
		if item.op.kind != txDelete && item.op.kind != txDeleteSymbolic {
			continue
		}

		deleted[item.target.name] = struct{}{}
		if item.target.ref.isPacked {
			needed = true
		}
	}

	if !needed {
		return nil
	}

	lock, err := tx.store.commonRoot.OpenFile("packed-refs.new", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}

	createdTemp := true

	defer func() {
		if !createdTemp {
			return
		}

		_ = tx.store.commonRoot.Remove("packed-refs.new")
	}()

	_, err = lock.WriteString("# pack-refs with: peeled fully-peeled sorted\n")
	if err != nil {
		_ = lock.Close()

		return err
	}

	for _, entry := range packed.ordered {
		if _, skip := deleted[entry.Name()]; skip {
			continue
		}

		_, err = lock.WriteString(entry.ID.String() + " " + entry.Name() + "\n")
		if err != nil {
			_ = lock.Close()

			return err
		}

		if entry.Peeled != nil {
			_, err = lock.WriteString("^" + entry.Peeled.String() + "\n")
			if err != nil {
				_ = lock.Close()

				return err
			}
		}
	}

	err = lock.Close()
	if err != nil {
		return err
	}

	err = tx.store.commonRoot.Rename("packed-refs.new", "packed-refs")
	if err != nil {
		return err
	}

	createdTemp = false

	return nil
}
