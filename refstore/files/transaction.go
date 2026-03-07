package files

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/ref/refname"
	"codeberg.org/lindenii/furgit/refstore"
)

type txKind uint8

const (
	txCreate txKind = iota
	txUpdate
	txDelete
	txVerify
	txCreateSymbolic
	txUpdateSymbolic
	txDeleteSymbolic
	txVerifySymbolic
)

type txOp struct {
	name      string
	kind      txKind
	newID     objectid.ObjectID
	oldID     objectid.ObjectID
	newTarget string
	oldTarget string
}

type directKind uint8

const (
	directMissing directKind = iota
	directDetached
	directSymbolic
)

type directRef struct {
	kind     directKind
	name     string
	id       objectid.ObjectID
	target   string
	isLoose  bool
	isPacked bool
}

type resolvedWriteTarget struct {
	name string
	loc  refPath
	ref  directRef
}

type preparedTxOp struct {
	op     txOp
	target resolvedWriteTarget
}

type Transaction struct {
	store  *Store
	ops    []txOp
	closed bool
}

var _ refstore.Transaction = (*Transaction)(nil)

// BeginTransaction creates one new files transaction.
//
//nolint:ireturn
func (store *Store) BeginTransaction() (refstore.Transaction, error) {
	return &Transaction{
		store: store,
		ops:   make([]txOp, 0, 8),
	}, nil
}

func (tx *Transaction) Create(name string, newID objectid.ObjectID) error {
	return tx.queue(txOp{name: name, kind: txCreate, newID: newID})
}

func (tx *Transaction) Update(name string, newID, oldID objectid.ObjectID) error {
	return tx.queue(txOp{name: name, kind: txUpdate, newID: newID, oldID: oldID})
}

func (tx *Transaction) Delete(name string, oldID objectid.ObjectID) error {
	return tx.queue(txOp{name: name, kind: txDelete, oldID: oldID})
}

func (tx *Transaction) Verify(name string, oldID objectid.ObjectID) error {
	return tx.queue(txOp{name: name, kind: txVerify, oldID: oldID})
}

func (tx *Transaction) CreateSymbolic(name, newTarget string) error {
	return tx.queue(txOp{name: name, kind: txCreateSymbolic, newTarget: newTarget})
}

func (tx *Transaction) UpdateSymbolic(name, newTarget, oldTarget string) error {
	return tx.queue(txOp{name: name, kind: txUpdateSymbolic, newTarget: newTarget, oldTarget: oldTarget})
}

func (tx *Transaction) DeleteSymbolic(name, oldTarget string) error {
	return tx.queue(txOp{name: name, kind: txDeleteSymbolic, oldTarget: oldTarget})
}

func (tx *Transaction) VerifySymbolic(name, oldTarget string) error {
	return tx.queue(txOp{name: name, kind: txVerifySymbolic, oldTarget: oldTarget})
}

func (tx *Transaction) Commit() error {
	err := tx.ensureOpen()
	if err != nil {
		return err
	}

	prepared, err := tx.prepare()
	if err != nil {
		tx.closed = true

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
			tx.closed = true

			return err
		}
	}

	err = tx.applyPackedDeletes(prepared)
	if err != nil {
		tx.closed = true

		return err
	}

	for _, item := range prepared {
		switch item.op.kind {
		case txDelete, txDeleteSymbolic:
			if item.target.ref.isLoose {
				err = tx.store.rootFor(item.target.loc.root).Remove(item.target.loc.path)
				if err != nil && !errors.Is(err, os.ErrNotExist) {
					tx.closed = true

					return err
				}

				tx.tryRemoveEmptyParents(item.target.name)
			}
		case txCreate, txUpdate, txVerify, txCreateSymbolic, txUpdateSymbolic, txVerifySymbolic:
		}
	}

	tx.closed = true

	return nil
}

func (tx *Transaction) Abort() error {
	err := tx.ensureOpen()
	if err != nil {
		return err
	}

	tx.closed = true

	return nil
}

func (tx *Transaction) ensureOpen() error {
	if tx.closed {
		return fmt.Errorf("refstore/files: transaction already closed")
	}

	return nil
}

func (tx *Transaction) queue(op txOp) error {
	err := tx.ensureOpen()
	if err != nil {
		return err
	}

	err = tx.validateOp(op)
	if err != nil {
		return err
	}

	tx.ops = append(tx.ops, op)

	return nil
}

func (tx *Transaction) validateOp(op txOp) error {
	if op.name == "" {
		return fmt.Errorf("refstore/files: empty reference name")
	}

	switch op.kind {
	case txCreate, txUpdate:
		err := refname.ValidateUpdateName(op.name, true)
		if err != nil {
			return err
		}

		if op.newID.Size() == 0 {
			return objectid.ErrInvalidAlgorithm
		}
	case txDelete, txVerify:
		err := refname.ValidateUpdateName(op.name, false)
		if err != nil {
			return err
		}

		if op.oldID.Size() == 0 {
			return objectid.ErrInvalidAlgorithm
		}
	case txCreateSymbolic, txUpdateSymbolic:
		err := refname.ValidateUpdateName(op.name, true)
		if err != nil {
			return err
		}

		if strings.TrimSpace(op.newTarget) == "" {
			return fmt.Errorf("refstore/files: empty symbolic target")
		}

		err = refname.ValidateSymbolicTarget(op.name, strings.TrimSpace(op.newTarget))
		if err != nil {
			return err
		}
	case txDeleteSymbolic, txVerifySymbolic:
		err := refname.ValidateUpdateName(op.name, false)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("refstore/files: unsupported transaction operation %d", op.kind)
	}

	if op.kind == txUpdateSymbolic || op.kind == txDeleteSymbolic || op.kind == txVerifySymbolic {
		if strings.TrimSpace(op.oldTarget) == "" {
			return fmt.Errorf("refstore/files: empty symbolic old target")
		}
	}

	return nil
}
