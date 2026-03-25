package files

import (
	"strings"

	"codeberg.org/lindenii/furgit/ref/store"
)

func (executor *refUpdateExecutor) verifyPreparedUpdateCurrent(item preparedUpdate) error {
	switch item.op.kind {
	case updateCreate:
		if item.target.ref.kind != directMissing {
			return wrapUpdateError(item.op.name, &refstore.CreateExistsError{})
		}

		return nil
	case updateReplace, updateDelete, updateVerify:
		if item.target.ref.kind == directMissing {
			return wrapUpdateError(item.op.name, refstore.ErrReferenceNotFound)
		}

		if item.target.ref.kind != directDetached {
			return wrapUpdateError(item.op.name, &refstore.ExpectedDetachedError{})
		}

		if item.target.ref.id != item.op.oldID {
			return wrapUpdateError(item.op.name, &refstore.IncorrectOldValueError{
				Actual:   item.target.ref.id.String(),
				Expected: item.op.oldID.String(),
			})
		}

		return nil
	case updateCreateSymbolic:
		if item.target.ref.kind != directMissing {
			return wrapUpdateError(item.op.name, &refstore.CreateExistsError{})
		}

		return nil
	case updateReplaceSymbolic, updateDeleteSymbolic, updateVerifySymbolic:
		if item.target.ref.kind == directMissing {
			return wrapUpdateError(item.op.name, refstore.ErrReferenceNotFound)
		}

		if item.target.ref.kind != directSymbolic {
			return wrapUpdateError(item.op.name, &refstore.ExpectedSymbolicError{})
		}

		if strings.TrimSpace(item.target.ref.target) != strings.TrimSpace(item.op.oldTarget) {
			return wrapUpdateError(item.op.name, &refstore.IncorrectOldValueError{
				Actual:   strings.TrimSpace(item.target.ref.target),
				Expected: strings.TrimSpace(item.op.oldTarget),
			})
		}

		return nil
	}

	return nil
}
