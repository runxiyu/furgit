package files

import (
	"fmt"
	"slices"
)

func (tx *Transaction) prepare() (prepared []preparedTxOp, err error) {
	prepared = make([]preparedTxOp, 0, len(tx.ops))

	defer func() {
		if err != nil {
			_ = tx.cleanup(prepared)
		}
	}()

	targets := make(map[string]struct{}, len(tx.ops))

	for _, op := range tx.ops {
		target, err := tx.resolveTarget(op)
		if err != nil {
			return prepared, err
		}

		targetKey := tx.targetKey(target.loc)
		if _, exists := targets[targetKey]; exists {
			return prepared, fmt.Errorf("refstore/files: duplicate transaction operation for %q", target.name)
		}

		targets[targetKey] = struct{}{}

		prepared = append(prepared, preparedTxOp{
			op:     op,
			target: target,
		})
	}

	deleted := make(map[string]struct{})
	written := make([]string, 0, len(prepared))

	for _, item := range prepared {
		switch item.op.kind {
		case txDelete, txDeleteSymbolic:
			deleted[item.target.name] = struct{}{}
		case txCreate, txUpdate, txCreateSymbolic, txUpdateSymbolic:
			written = append(written, item.target.name)
		case txVerify, txVerifySymbolic:
		}
	}

	existing, err := tx.visibleNames()
	if err != nil {
		return prepared, err
	}

	for _, name := range written {
		err = verifyRefnameAvailable(name, existing, written, deleted)
		if err != nil {
			return prepared, err
		}
	}

	lockNames := make([]string, 0, len(prepared))
	for _, item := range prepared {
		lockNames = append(lockNames, tx.targetKey(item.target.loc))
	}

	slices.Sort(lockNames)

	for _, lockKey := range lockNames {
		err = tx.createLock(refPathFromKey(lockKey))
		if err != nil {
			return prepared, err
		}
	}

	hasDeletes := len(deleted) > 0
	if hasDeletes {
		err = tx.createPackedLock(tx.store.packedRefsTimeout)
		if err != nil {
			return prepared, err
		}
	}

	for i := range prepared {
		item := &prepared[i]

		refState, err := tx.directRead(item.target.name)
		if err != nil {
			return prepared, err
		}

		item.target.ref = refState

		err = tx.verifyCurrent(*item)
		if err != nil {
			return prepared, err
		}
	}

	return prepared, nil
}
