package files

import refstore "lindenii.org/go/furgit/ref/store"

func (executor *refUpdateExecutor) resolvePreparedUpdates(ops []queuedUpdate) ([]preparedUpdate, error) {
	prepared := make([]preparedUpdate, 0, len(ops))
	targets := make(map[string]struct{}, len(ops))

	for _, op := range ops {
		target, err := executor.resolveQueuedUpdateTarget(op)
		if err != nil {
			return prepared, err
		}

		targetKey := updateTargetKey(target.loc)
		if _, exists := targets[targetKey]; exists {
			return prepared, wrapUpdateError(op.name, &refstore.DuplicateUpdateError{})
		}

		targets[targetKey] = struct{}{}

		prepared = append(prepared, preparedUpdate{op: op, target: target})
	}

	return prepared, nil
}

func collectPreparedWrites(prepared []preparedUpdate) (deleted map[string]struct{}, written []string) {
	deleted = make(map[string]struct{})
	written = make([]string, 0, len(prepared))

	for _, item := range prepared {
		switch item.op.kind {
		case updateDelete, updateDeleteSymbolic:
			deleted[item.target.name] = struct{}{}
		case updateCreate, updateReplace, updateCreateSymbolic, updateReplaceSymbolic:
			written = append(written, item.target.name)
		case updateVerify, updateVerifySymbolic:
		}
	}

	return deleted, written
}
