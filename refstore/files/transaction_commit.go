package files

import (
	"errors"
	"os"
)

func (tx *Transaction) Commit() error {
	prepared, err := tx.prepare()
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.cleanup(prepared)
	}()

	for _, item := range prepared {
		if item.op.kind == txDelete || item.op.kind == txDeleteSymbolic || item.op.kind == txVerify || item.op.kind == txVerifySymbolic {
			continue
		}

		err = tx.writeLoose(item)
		if err != nil {
			return err
		}
	}

	err = tx.applyPackedDeletes(prepared)
	if err != nil {
		return err
	}

	for _, item := range prepared {
		switch item.op.kind {
		case txDelete, txDeleteSymbolic:
			if item.target.ref.isLoose {
				err = tx.store.rootFor(item.target.loc.root).Remove(item.target.loc.path)
				if err != nil && !errors.Is(err, os.ErrNotExist) {
					return err
				}

				tx.tryRemoveEmptyParents(item.target.name)
			}
		case txCreate, txUpdate, txVerify, txCreateSymbolic, txUpdateSymbolic, txVerifySymbolic:
		}
	}

	return nil
}
