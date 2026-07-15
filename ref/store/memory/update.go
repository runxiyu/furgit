package memory

import (
	"errors"
	"fmt"

	"lindenii.org/go/furgit/object/id"
	refname "lindenii.org/go/furgit/ref/name"
	"lindenii.org/go/furgit/ref/store"
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
	newID     id.ObjectID //exhaustruct:optional
	oldID     id.ObjectID //exhaustruct:optional
	newTarget string      //exhaustruct:optional
	oldTarget string      //exhaustruct:optional
}

type resolvedUpdateTarget struct {
	name string
	ref  storedRef
}

type preparedUpdate struct {
	op     queuedUpdate
	target resolvedUpdateTarget
}

// validateQueuedUpdate checks one operation at queue time,
// rejecting malformed names and values
// before they can enter a transaction or batch.
func validateQueuedUpdate(objectFormat id.ObjectFormat, op queuedUpdate) error {
	switch op.kind {
	case updateCreate, updateReplace:
		err := refname.ValidateUpdateName(op.name, true)
		if err != nil {
			return fmt.Errorf("ref/store/memory: %w", err)
		}

		if op.newID.ObjectFormat() != objectFormat {
			return fmt.Errorf("%w: object id format mismatch", store.ErrInvalidValue)
		}
	case updateDelete, updateVerify:
		err := refname.ValidateUpdateName(op.name, false)
		if err != nil {
			return fmt.Errorf("ref/store/memory: %w", err)
		}

		if op.oldID.ObjectFormat() != objectFormat {
			return fmt.Errorf("%w: object id format mismatch", store.ErrInvalidValue)
		}
	case updateCreateSymbolic, updateReplaceSymbolic:
		err := refname.ValidateUpdateName(op.name, true)
		if err != nil {
			return fmt.Errorf("ref/store/memory: %w", err)
		}

		if op.newTarget == "" {
			return fmt.Errorf("%w: empty symbolic target", store.ErrInvalidValue)
		}

		err = refname.ValidateSymbolicTarget(op.name, op.newTarget)
		if err != nil {
			return fmt.Errorf("ref/store/memory: %w", err)
		}
	case updateDeleteSymbolic, updateVerifySymbolic:
		err := refname.ValidateUpdateName(op.name, false)
		if err != nil {
			return fmt.Errorf("ref/store/memory: %w", err)
		}
	default:
		panic(fmt.Sprintf("ref/store/memory: unsupported update operation %d", op.kind))
	}

	if op.kind == updateReplaceSymbolic || op.kind == updateDeleteSymbolic || op.kind == updateVerifySymbolic {
		if op.oldTarget == "" {
			return fmt.Errorf("%w: empty symbolic old target", store.ErrInvalidValue)
		}
	}

	return nil
}

// prepareUpdates resolves, conflict-checks, and verifies a queued operation
// set against refs without mutating it.
// On failure it returns the name of the offending operation alongside the error.
func prepareUpdates(refs map[string]storedRef, ops []queuedUpdate) ([]preparedUpdate, string, error) {
	prepared, name, err := resolvePreparedUpdates(refs, ops)
	if err != nil {
		return prepared, name, err
	}

	deleted, written := collectPreparedWrites(prepared)
	existing := collectVisibleNames(refs)

	for _, name := range written {
		err = verifyRefnameAvailable(name, existing, written, deleted)
		if err != nil {
			return prepared, name, err
		}
	}

	name, err = verifyPreparedUpdates(refs, prepared)
	if err != nil {
		return prepared, name, err
	}

	return prepared, "", nil
}

func resolvePreparedUpdates(refs map[string]storedRef, ops []queuedUpdate) ([]preparedUpdate, string, error) {
	prepared := make([]preparedUpdate, 0, len(ops))
	targets := make(map[string]struct{}, len(ops))

	for _, op := range ops {
		target, err := resolveQueuedUpdateTarget(refs, op)
		if err != nil {
			return prepared, op.name, err
		}

		if _, exists := targets[target.name]; exists {
			return prepared, op.name, store.ErrDuplicateUpdate
		}

		targets[target.name] = struct{}{}
		prepared = append(prepared, preparedUpdate{op: op, target: target})
	}

	return prepared, "", nil
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
		panic(fmt.Sprintf("ref/store/memory: unsupported update operation %d", op.kind))
	}
}

func resolveOrdinaryTarget(refs map[string]storedRef, name string, allowMissing bool) (resolvedUpdateTarget, error) {
	cur := name
	seen := make(map[string]struct{})

	for {
		if _, ok := seen[cur]; ok {
			return resolvedUpdateTarget{}, fmt.Errorf("%w: at %q", store.ErrSymbolicCycle, cur)
		}

		seen[cur] = struct{}{}

		refState := directRead(refs, cur)
		switch refState.kind {
		case storedMissing:
			if !allowMissing {
				return resolvedUpdateTarget{}, store.ErrReferenceNotFound
			}

			return resolvedUpdateTarget{name: cur, ref: refState}, nil
		case storedDirect:
			return resolvedUpdateTarget{name: cur, ref: refState}, nil
		case storedSymbolic:
			if refState.target == "" {
				return resolvedUpdateTarget{}, fmt.Errorf(
					"%w: symbolic reference has empty target", store.ErrInvalidValue,
				)
			}

			cur = refState.target
		default:
			panic(fmt.Sprintf("ref/store/memory: unsupported stored reference kind %d", refState.kind))
		}
	}
}

func directRead(refs map[string]storedRef, name string) storedRef {
	stored, ok := refs[name]
	if !ok {
		return storedRef{kind: storedMissing}
	}

	return stored
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
		default:
			panic(fmt.Sprintf("ref/store/memory: unsupported update operation %d", item.op.kind))
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
			return &store.NameConflictError{Other: existingName}
		}
	}

	for _, other := range writes {
		if other == name {
			continue
		}

		if refnamesConflict(name, other) {
			return &store.NameConflictError{Other: other}
		}
	}

	return nil
}

func refnamesConflict(left, right string) bool {
	return left == right ||
		hasPathPrefix(left, right) ||
		hasPathPrefix(right, left)
}

func hasPathPrefix(name, prefix string) bool {
	return len(name) > len(prefix) &&
		name[len(prefix)] == '/' &&
		name[:len(prefix)] == prefix
}

func verifyPreparedUpdates(refs map[string]storedRef, prepared []preparedUpdate) (string, error) {
	for i := range prepared {
		item := &prepared[i]
		item.target.ref = directRead(refs, item.target.name)

		err := verifyPreparedUpdateCurrent(*item)
		if err != nil {
			return item.op.name, err
		}
	}

	return "", nil
}

func verifyPreparedUpdateCurrent(item preparedUpdate) error {
	switch item.op.kind {
	case updateCreate, updateCreateSymbolic:
		if item.target.ref.kind != storedMissing {
			return store.ErrCreateExists
		}

		return nil
	case updateReplace, updateDelete, updateVerify:
		if item.target.ref.kind == storedMissing {
			return store.ErrReferenceNotFound
		}

		if item.target.ref.kind != storedDirect {
			return store.ErrExpectedDirect
		}

		if item.target.ref.id != item.op.oldID {
			return &store.WrongOldIDError{Actual: item.target.ref.id, Expected: item.op.oldID}
		}

		return nil
	case updateReplaceSymbolic, updateDeleteSymbolic, updateVerifySymbolic:
		if item.target.ref.kind == storedMissing {
			return store.ErrReferenceNotFound
		}

		if item.target.ref.kind != storedSymbolic {
			return store.ErrExpectedSymbolic
		}

		if item.target.ref.target != item.op.oldTarget {
			return &store.WrongOldTargetError{Actual: item.target.ref.target, Expected: item.op.oldTarget}
		}

		return nil
	default:
		panic(fmt.Sprintf("ref/store/memory: unsupported update operation %d", item.op.kind))
	}
}

func applyPreparedUpdates(refs map[string]storedRef, prepared []preparedUpdate) {
	for _, item := range prepared {
		switch item.op.kind {
		case updateCreate, updateReplace:
			refs[item.target.name] = storedRef{kind: storedDirect, id: item.op.newID}
		case updateCreateSymbolic, updateReplaceSymbolic:
			refs[item.target.name] = storedRef{kind: storedSymbolic, target: item.op.newTarget}
		case updateDelete, updateDeleteSymbolic:
			delete(refs, item.target.name)
		case updateVerify, updateVerifySymbolic:
		default:
			panic(fmt.Sprintf("ref/store/memory: unsupported update operation %d", item.op.kind))
		}
	}
}

// isBatchRejected reports whether err is a per-operation rejection
// that should drop only the offending operation,
// rather than a fatal failure that aborts the whole batch.
func isBatchRejected(err error) bool {
	switch {
	case errors.Is(err, store.ErrReferenceNotFound),
		errors.Is(err, store.ErrCreateExists),
		errors.Is(err, store.ErrDuplicateUpdate),
		errors.Is(err, store.ErrExpectedDirect),
		errors.Is(err, store.ErrExpectedSymbolic),
		errors.Is(err, store.ErrInvalidValue),
		errors.Is(err, store.ErrSymbolicCycle),
		errors.Is(err, refname.ErrInvalidName):
		return true
	}

	if _, ok := errors.AsType[*store.NameConflictError](err); ok {
		return true
	}

	if _, ok := errors.AsType[*store.WrongOldIDError](err); ok {
		return true
	}

	if _, ok := errors.AsType[*store.WrongOldTargetError](err); ok {
		return true
	}

	return false
}
