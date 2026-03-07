package files

import "fmt"

func (tx *Transaction) resolveTarget(op txOp) (resolvedWriteTarget, error) {
	switch op.kind {
	case txCreate:
		return tx.resolveOrdinaryTarget(op.name, true)
	case txUpdate, txDelete, txVerify:
		return tx.resolveOrdinaryTarget(op.name, false)
	case txCreateSymbolic, txUpdateSymbolic, txDeleteSymbolic, txVerifySymbolic:
		refState, err := tx.directRead(op.name)
		if err != nil {
			return resolvedWriteTarget{}, err
		}

		return resolvedWriteTarget{name: op.name, loc: tx.store.loosePath(op.name), ref: refState}, nil
	default:
		return resolvedWriteTarget{}, fmt.Errorf("refstore/files: unsupported transaction operation %d", op.kind)
	}
}
