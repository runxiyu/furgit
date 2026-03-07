package files

import (
	"errors"
	"os"
)

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
