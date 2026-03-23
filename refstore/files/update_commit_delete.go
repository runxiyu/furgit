package files

import (
	"errors"
	"os"
)

func (executor *refUpdateExecutor) removeDeletedLooseRefs(prepared []preparedUpdate) error {
	for _, item := range prepared {
		switch item.op.kind {
		case updateDelete, updateDeleteSymbolic:
			if item.target.ref.isLoose {
				err := executor.store.rootFor(item.target.loc.root).Remove(item.target.loc.path)
				if err != nil && !errors.Is(err, os.ErrNotExist) {
					return wrapUpdateError(item.op.name, err)
				}

				executor.tryRemoveEmptyParents(item.target.name)
			}
		case updateCreate, updateReplace, updateVerify, updateCreateSymbolic, updateReplaceSymbolic, updateVerifySymbolic:
		}
	}

	return nil
}
