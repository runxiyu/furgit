package files

import (
	"fmt"
	"strings"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/ref/refname"
	refstore "codeberg.org/lindenii/furgit/ref/store"
)

func (executor *refUpdateExecutor) validateQueuedUpdate(op queuedUpdate) error {
	if op.name == "" {
		return wrapUpdateError(op.name, &refstore.InvalidNameError{Err: fmt.Errorf("empty reference name")})
	}

	switch op.kind {
	case updateCreate, updateReplace:
		err := refname.ValidateUpdateName(op.name, true)
		if err != nil {
			return wrapUpdateError(op.name, &refstore.InvalidNameError{Err: err})
		}

		if op.newID.Size() == 0 {
			return wrapUpdateError(op.name, &refstore.InvalidValueError{Err: objectid.ErrInvalidAlgorithm})
		}
	case updateDelete, updateVerify:
		err := refname.ValidateUpdateName(op.name, false)
		if err != nil {
			return wrapUpdateError(op.name, &refstore.InvalidNameError{Err: err})
		}

		if op.oldID.Size() == 0 {
			return wrapUpdateError(op.name, &refstore.InvalidValueError{Err: objectid.ErrInvalidAlgorithm})
		}
	case updateCreateSymbolic, updateReplaceSymbolic:
		err := refname.ValidateUpdateName(op.name, true)
		if err != nil {
			return wrapUpdateError(op.name, &refstore.InvalidNameError{Err: err})
		}

		if strings.TrimSpace(op.newTarget) == "" {
			return wrapUpdateError(op.name, &refstore.InvalidValueError{Err: fmt.Errorf("empty symbolic target")})
		}

		err = refname.ValidateSymbolicTarget(op.name, strings.TrimSpace(op.newTarget))
		if err != nil {
			return wrapUpdateError(op.name, &refstore.InvalidValueError{Err: err})
		}
	case updateDeleteSymbolic, updateVerifySymbolic:
		err := refname.ValidateUpdateName(op.name, false)
		if err != nil {
			return wrapUpdateError(op.name, &refstore.InvalidNameError{Err: err})
		}
	default:
		return fmt.Errorf("refstore/files: unsupported update operation %d", op.kind)
	}

	if op.kind == updateReplaceSymbolic || op.kind == updateDeleteSymbolic || op.kind == updateVerifySymbolic {
		if strings.TrimSpace(op.oldTarget) == "" {
			return wrapUpdateError(op.name, &refstore.InvalidValueError{Err: fmt.Errorf("empty symbolic old target")})
		}
	}

	return nil
}
