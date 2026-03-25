package files

import "fmt"

func (executor *refUpdateExecutor) resolveQueuedUpdateTarget(op queuedUpdate) (resolvedUpdateTarget, error) {
	switch op.kind {
	case updateCreate:
		return executor.resolveOrdinaryTarget(op.name, true)
	case updateReplace, updateDelete, updateVerify:
		return executor.resolveOrdinaryTarget(op.name, false)
	case updateCreateSymbolic, updateReplaceSymbolic, updateDeleteSymbolic, updateVerifySymbolic:
		refState, err := executor.directRead(op.name)
		if err != nil {
			return resolvedUpdateTarget{}, err
		}

		return resolvedUpdateTarget{name: op.name, loc: executor.store.loosePath(op.name), ref: refState}, nil
	default:
		return resolvedUpdateTarget{}, fmt.Errorf("refstore/files: unsupported update operation %d", op.kind)
	}
}
