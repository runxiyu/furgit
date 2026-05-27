package memory

import (
	"errors"
	"fmt"
	"strings"

	objectid "lindenii.org/go/furgit/object/id"
	refname "lindenii.org/go/furgit/ref/name"
	refstore "lindenii.org/go/furgit/ref/store"
)

type updateKind uint8

const (
	updateCreate updateKind = iota
	updateReplace
	updateDelete
	updateVerify
	updateCreateSymbolic
	updateReplaceSymbolic
	updateDeleteSymbolic
	updateVerifySymbolic
)

type queuedUpdate struct {
	name      string
	kind      updateKind
	newID     objectid.ObjectID
	oldID     objectid.ObjectID
	newTarget string
	oldTarget string
}

type resolvedUpdateTarget struct {
	name string
	ref  storedRef
}

type preparedUpdate struct {
	op     queuedUpdate
	target resolvedUpdateTarget
}

type updateContextError struct {
	name string
	err  error
}

func (err *updateContextError) Error() string {
	return fmt.Sprintf("refstore/memory: update %q: %v", err.name, err.err)
}

func (err *updateContextError) Unwrap() error {
	if err == nil {
		return nil
	}

	return err.err
}

func wrapUpdateError(name string, err error) error {
	if err == nil || name == "" {
		return err
	}

	return &updateContextError{name: name, err: err}
}

func validateQueuedUpdate(op queuedUpdate) error {
	if op.name == "" {
		return wrapUpdateError(op.name, &refstore.InvalidNameError{Err: fmt.Errorf("empty reference name")})
	}

	switch op.kind {
	case updateCreate, updateReplace:
		err := refname.ValidateUpdateName(op.name, true)
		if err != nil {
			return wrapUpdateError(op.name, &refstore.InvalidNameError{Err: err})
		}

		if op.newID.Algorithm().Size() == 0 {
			return wrapUpdateError(op.name, &refstore.InvalidValueError{Err: objectid.ErrInvalidAlgorithm})
		}
	case updateDelete, updateVerify:
		err := refname.ValidateUpdateName(op.name, false)
		if err != nil {
			return wrapUpdateError(op.name, &refstore.InvalidNameError{Err: err})
		}

		if op.oldID.Algorithm().Size() == 0 {
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
		return fmt.Errorf("refstore/memory: unsupported update operation %d", op.kind)
	}

	if op.kind == updateReplaceSymbolic || op.kind == updateDeleteSymbolic || op.kind == updateVerifySymbolic {
		if strings.TrimSpace(op.oldTarget) == "" {
			return wrapUpdateError(op.name, &refstore.InvalidValueError{Err: fmt.Errorf("empty symbolic old target")})
		}
	}

	return nil
}

func prepareUpdates(refs map[string]storedRef, ops []queuedUpdate) ([]preparedUpdate, error) {
	prepared, err := resolvePreparedUpdates(refs, ops)
	if err != nil {
		return prepared, err
	}

	deleted, written := collectPreparedWrites(prepared)
	existing := collectVisibleNames(refs)

	for _, name := range written {
		err = verifyRefnameAvailable(name, existing, written, deleted)
		if err != nil {
			return prepared, err
		}
	}

	err = verifyPreparedUpdates(refs, prepared)
	if err != nil {
		return prepared, err
	}

	return prepared, nil
}

func resolvePreparedUpdates(refs map[string]storedRef, ops []queuedUpdate) ([]preparedUpdate, error) {
	prepared := make([]preparedUpdate, 0, len(ops))
	targets := make(map[string]struct{}, len(ops))

	for _, op := range ops {
		target, err := resolveQueuedUpdateTarget(refs, op)
		if err != nil {
			return prepared, err
		}

		if _, exists := targets[target.name]; exists {
			return prepared, wrapUpdateError(op.name, &refstore.DuplicateUpdateError{})
		}

		targets[target.name] = struct{}{}
		prepared = append(prepared, preparedUpdate{op: op, target: target})
	}

	return prepared, nil
}

func resolveQueuedUpdateTarget(refs map[string]storedRef, op queuedUpdate) (resolvedUpdateTarget, error) {
	switch op.kind {
	case updateCreate:
		return resolveOrdinaryTarget(refs, op.name, true)
	case updateReplace, updateDelete, updateVerify:
		return resolveOrdinaryTarget(refs, op.name, false)
	case updateCreateSymbolic, updateReplaceSymbolic, updateDeleteSymbolic, updateVerifySymbolic:
		return resolvedUpdateTarget{name: op.name, ref: directRead(refs, op.name)}, nil
	default:
		return resolvedUpdateTarget{}, fmt.Errorf("refstore/memory: unsupported update operation %d", op.kind)
	}
}

func resolveOrdinaryTarget(refs map[string]storedRef, name string, allowMissing bool) (resolvedUpdateTarget, error) {
	cur := name
	seen := make(map[string]struct{})

	for {
		if _, ok := seen[cur]; ok {
			return resolvedUpdateTarget{}, fmt.Errorf("refstore/memory: symbolic reference cycle at %q", cur)
		}

		seen[cur] = struct{}{}

		refState := directRead(refs, cur)
		switch refState.kind {
		case storedMissing:
			if !allowMissing {
				return resolvedUpdateTarget{}, wrapUpdateError(name, refstore.ErrReferenceNotFound)
			}

			return resolvedUpdateTarget{name: cur, ref: refState}, nil
		case storedDetached:
			return resolvedUpdateTarget{name: cur, ref: refState}, nil
		case storedSymbolic:
			target := strings.TrimSpace(refState.target)
			if target == "" {
				return resolvedUpdateTarget{}, wrapUpdateError(name, &refstore.InvalidValueError{
					Err: fmt.Errorf("symbolic reference has empty target"),
				})
			}

			cur = target
		default:
			return resolvedUpdateTarget{}, fmt.Errorf("refstore/memory: unsupported stored reference kind %d", refState.kind)
		}
	}
}

func directRead(refs map[string]storedRef, name string) storedRef {
	stored, ok := refs[name]
	if !ok {
		return storedRef{kind: storedMissing}
	}

	return cloneStoredRef(stored)
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

func collectVisibleNames(refs map[string]storedRef) map[string]struct{} {
	names := make(map[string]struct{}, len(refs))
	for name := range refs {
		names[name] = struct{}{}
	}

	return names
}

func verifyRefnameAvailable(name string, existing map[string]struct{}, writes []string, deleted map[string]struct{}) error {
	for existingName := range existing {
		if existingName == name {
			continue
		}

		if _, skip := deleted[existingName]; skip {
			continue
		}

		if refnamesConflict(name, existingName) {
			return wrapUpdateError(name, &refstore.NameConflictError{Other: existingName})
		}
	}

	for _, other := range writes {
		if other == name {
			continue
		}

		if refnamesConflict(name, other) {
			return wrapUpdateError(name, &refstore.NameConflictError{Other: other})
		}
	}

	return nil
}

func refnamesConflict(left, right string) bool {
	return left == right ||
		strings.HasPrefix(left, right+"/") ||
		strings.HasPrefix(right, left+"/")
}

func verifyPreparedUpdates(refs map[string]storedRef, prepared []preparedUpdate) error {
	for i := range prepared {
		item := &prepared[i]
		item.target.ref = directRead(refs, item.target.name)

		err := verifyPreparedUpdateCurrent(*item)
		if err != nil {
			return err
		}
	}

	return nil
}

func verifyPreparedUpdateCurrent(item preparedUpdate) error {
	switch item.op.kind {
	case updateCreate:
		if item.target.ref.kind != storedMissing {
			return wrapUpdateError(item.op.name, &refstore.CreateExistsError{})
		}

		return nil
	case updateReplace, updateDelete, updateVerify:
		if item.target.ref.kind == storedMissing {
			return wrapUpdateError(item.op.name, refstore.ErrReferenceNotFound)
		}

		if item.target.ref.kind != storedDetached {
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
		if item.target.ref.kind != storedMissing {
			return wrapUpdateError(item.op.name, &refstore.CreateExistsError{})
		}

		return nil
	case updateReplaceSymbolic, updateDeleteSymbolic, updateVerifySymbolic:
		if item.target.ref.kind == storedMissing {
			return wrapUpdateError(item.op.name, refstore.ErrReferenceNotFound)
		}

		if item.target.ref.kind != storedSymbolic {
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

func applyPreparedUpdates(refs map[string]storedRef, prepared []preparedUpdate) {
	for _, item := range prepared {
		switch item.op.kind {
		case updateCreate, updateReplace:
			refs[item.target.name] = storedRef{kind: storedDetached, id: item.op.newID}
		case updateCreateSymbolic, updateReplaceSymbolic:
			refs[item.target.name] = storedRef{kind: storedSymbolic, target: strings.TrimSpace(item.op.newTarget)}
		case updateDelete, updateDeleteSymbolic:
			delete(refs, item.target.name)
		case updateVerify, updateVerifySymbolic:
		}
	}
}

func isBatchRejected(err error) bool {
	_, invalidName := errors.AsType[*refstore.InvalidNameError](err)
	_, invalidValue := errors.AsType[*refstore.InvalidValueError](err)
	_, duplicateUpdate := errors.AsType[*refstore.DuplicateUpdateError](err)
	_, createExists := errors.AsType[*refstore.CreateExistsError](err)
	_, incorrectOldValue := errors.AsType[*refstore.IncorrectOldValueError](err)
	_, expectedDetached := errors.AsType[*refstore.ExpectedDetachedError](err)
	_, expectedSymbolic := errors.AsType[*refstore.ExpectedSymbolicError](err)
	_, nameConflict := errors.AsType[*refstore.NameConflictError](err)

	return errors.Is(err, refstore.ErrReferenceNotFound) ||
		invalidName ||
		invalidValue ||
		duplicateUpdate ||
		createExists ||
		incorrectOldValue ||
		expectedDetached ||
		expectedSymbolic ||
		nameConflict
}

func batchResultError(err error) error {
	updateErr, ok := errors.AsType[*updateContextError](err)
	if ok {
		return updateErr.err
	}

	return err
}

func batchResultName(err error) string {
	updateErr, ok := errors.AsType[*updateContextError](err)
	if !ok {
		return ""
	}

	return updateErr.name
}

// TODO: these really shouldn't be used in the batched path, really...
