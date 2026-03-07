package files

import (
	"fmt"
	"strings"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/ref/refname"
)

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
