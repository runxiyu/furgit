package files

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref"
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

type directRefKind uint8

const (
	directMissing directRefKind = iota
	directDirect
	directSymbolic
)

// directRefState is the current on-disk state of one reference name,
// read without following symbolic references.
type directRefState struct {
	kind     directRefKind
	id       id.ObjectID //exhaustruct:optional
	target   string      //exhaustruct:optional
	isLoose  bool        //exhaustruct:optional
	isPacked bool        //exhaustruct:optional
}

type resolvedUpdateTarget struct {
	name string
	loc  refPath
	ref  directRefState
}

type preparedUpdate struct {
	op     queuedUpdate
	target resolvedUpdateTarget
}

// updateExecutor carries the lock state of one prepare/commit cycle.
type updateExecutor struct {
	files        *Files
	lockedRefs   []refPath //exhaustruct:optional
	packedLocked bool      //exhaustruct:optional

	// deletedNames are deleted reference names
	// whose empty parent directories are pruned during cleanup,
	// after their lock files are gone.
	deletedNames []string //exhaustruct:optional
}

// validateQueuedUpdate checks one operation at queue time,
// rejecting malformed names and values
// before they can enter a transaction or batch.
func validateQueuedUpdate(objectFormat id.ObjectFormat, op queuedUpdate) error {
	switch op.kind {
	case updateCreate, updateReplace:
		err := refname.ValidateUpdateName(op.name, true)
		if err != nil {
			return fmt.Errorf("ref/store/files: %w", err)
		}

		if op.newID.ObjectFormat() != objectFormat {
			return fmt.Errorf("%w: object id format mismatch", store.ErrInvalidValue)
		}
	case updateDelete, updateVerify:
		err := refname.ValidateUpdateName(op.name, false)
		if err != nil {
			return fmt.Errorf("ref/store/files: %w", err)
		}

		if op.oldID.ObjectFormat() != objectFormat {
			return fmt.Errorf("%w: object id format mismatch", store.ErrInvalidValue)
		}
	case updateCreateSymbolic, updateReplaceSymbolic:
		err := refname.ValidateUpdateName(op.name, true)
		if err != nil {
			return fmt.Errorf("ref/store/files: %w", err)
		}

		if op.newTarget == "" {
			return fmt.Errorf("%w: empty symbolic target", store.ErrInvalidValue)
		}

		err = refname.ValidateSymbolicTarget(op.name, op.newTarget)
		if err != nil {
			return fmt.Errorf("ref/store/files: %w", err)
		}
	case updateDeleteSymbolic, updateVerifySymbolic:
		err := refname.ValidateUpdateName(op.name, false)
		if err != nil {
			return fmt.Errorf("ref/store/files: %w", err)
		}
	default:
		panic(fmt.Sprintf("ref/store/files: unsupported update operation %d", op.kind))
	}

	if op.kind == updateReplaceSymbolic || op.kind == updateDeleteSymbolic || op.kind == updateVerifySymbolic {
		if op.oldTarget == "" {
			return fmt.Errorf("%w: empty symbolic old target", store.ErrInvalidValue)
		}
	}

	return nil
}

// prepare resolves, conflict-checks, locks, and verifies
// one queued operation set.
// On failure it returns the name of the offending operation
// alongside the error,
// leaving any acquired locks for the caller to release.
func (executor *updateExecutor) prepare(ops []queuedUpdate) ([]preparedUpdate, string, error) {
	prepared, name, err := executor.resolveUpdates(ops)
	if err != nil {
		return prepared, name, err
	}

	deleted, written := collectPreparedWrites(prepared)

	existing, err := executor.collectVisibleNames()
	if err != nil {
		return prepared, "", err
	}

	for _, name := range written {
		err = verifyRefnameAvailable(name, existing, written, deleted)
		if err != nil {
			return prepared, name, err
		}
	}

	name, err = executor.lockPrepared(prepared)
	if err != nil {
		return prepared, name, err
	}

	if len(deleted) > 0 {
		err = executor.lockPackedRefs()
		if err != nil {
			return prepared, "", err
		}
	}

	name, err = executor.verifyPrepared(prepared)
	if err != nil {
		return prepared, name, err
	}

	return prepared, "", nil
}

func (executor *updateExecutor) resolveUpdates(ops []queuedUpdate) ([]preparedUpdate, string, error) {
	prepared := make([]preparedUpdate, 0, len(ops))
	targets := make(map[refPath]struct{}, len(ops))

	for _, op := range ops {
		target, err := executor.resolveQueuedUpdateTarget(op)
		if err != nil {
			return prepared, op.name, err
		}

		if _, exists := targets[target.loc]; exists {
			return prepared, op.name, store.ErrDuplicateUpdate
		}

		targets[target.loc] = struct{}{}
		prepared = append(prepared, preparedUpdate{op: op, target: target})
	}

	return prepared, "", nil
}

func (executor *updateExecutor) resolveQueuedUpdateTarget(op queuedUpdate) (resolvedUpdateTarget, error) {
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

		return resolvedUpdateTarget{
			name: op.name,
			loc:  executor.files.loosePath(op.name),
			ref:  refState,
		}, nil
	default:
		panic(fmt.Sprintf("ref/store/files: unsupported update operation %d", op.kind))
	}
}

// resolveOrdinaryTarget follows symbolic references from name
// to the reference an ordinary operation applies to.
func (executor *updateExecutor) resolveOrdinaryTarget(name string, allowMissing bool) (resolvedUpdateTarget, error) {
	cur := name
	seen := make(map[string]struct{})

	for {
		if _, ok := seen[cur]; ok {
			return resolvedUpdateTarget{}, fmt.Errorf("%w: at %q", store.ErrSymbolicCycle, cur)
		}

		seen[cur] = struct{}{}

		refState, err := executor.directRead(cur)
		if err != nil {
			return resolvedUpdateTarget{}, err
		}

		switch refState.kind {
		case directMissing:
			if !allowMissing {
				return resolvedUpdateTarget{}, store.ErrReferenceNotFound
			}

			return resolvedUpdateTarget{name: cur, loc: executor.files.loosePath(cur), ref: refState}, nil
		case directDirect:
			return resolvedUpdateTarget{name: cur, loc: executor.files.loosePath(cur), ref: refState}, nil
		case directSymbolic:
			if refState.target == "" {
				return resolvedUpdateTarget{}, fmt.Errorf(
					"%w: symbolic reference has empty target", store.ErrInvalidValue,
				)
			}

			cur = refState.target
		default:
			panic(fmt.Sprintf("ref/store/files: unsupported direct reference state %d", refState.kind))
		}
	}
}

// directRead reads the current state of one reference name
// without following symbolic references,
// recording whether loose and packed versions exist.
func (executor *updateExecutor) directRead(name string) (directRefState, error) {
	hasPacked := false

	var packedEntry ref.Direct

	if refname.ParseWorktree(name).Type == refname.WorktreeShared {
		packed, err := executor.files.readPackedRefs()
		if err != nil {
			return directRefState{}, err
		}

		packedEntry, hasPacked = packed.byName[name]
	}

	loose, err := executor.files.readLooseRef(name)
	if err == nil {
		switch loose := loose.(type) {
		case ref.Direct:
			return directRefState{
				kind:     directDirect,
				id:       loose.ID,
				isLoose:  true,
				isPacked: hasPacked,
			}, nil
		case ref.Symbolic:
			return directRefState{
				kind:     directSymbolic,
				target:   loose.Target,
				isLoose:  true,
				isPacked: hasPacked,
			}, nil
		default:
			panic(fmt.Sprintf("ref/store/files: unsupported reference type %T", loose))
		}
	}

	if !errors.Is(err, store.ErrReferenceNotFound) && !errors.Is(err, errRefDirectory) {
		return directRefState{}, err
	}

	if hasPacked {
		return directRefState{kind: directDirect, id: packedEntry.ID, isPacked: true}, nil
	}

	return directRefState{kind: directMissing}, nil
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
			panic(fmt.Sprintf("ref/store/files: unsupported update operation %d", item.op.kind))
		}
	}

	return deleted, written
}

// collectVisibleNames collects every visible reference name,
// loose and packed,
// for name conflict checking.
func (executor *updateExecutor) collectVisibleNames() (map[string]struct{}, error) {
	names := make(map[string]struct{})

	_, err := executor.files.gitRoot.Stat("HEAD")
	if err == nil {
		names["HEAD"] = struct{}{}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("ref/store/files: stat HEAD: %w", err)
	}

	err = collectLooseRefNames(executor.files.commonRoot, "refs", names)
	if err != nil {
		return nil, err
	}

	err = collectLooseRefNames(executor.files.gitRoot, "refs", names)
	if err != nil {
		return nil, err
	}

	packed, err := executor.files.readPackedRefs()
	if err != nil {
		return nil, err
	}

	for name := range packed.byName {
		names[name] = struct{}{}
	}

	return names, nil
}

// collectLooseRefNames walks one loose reference directory tree,
// adding file names to names and skipping lock files.
func collectLooseRefNames(root *os.Root, dir string, names map[string]struct{}) error {
	file, err := root.Open(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("ref/store/files: open %q: %w", dir, err)
	}

	entries, err := file.ReadDir(-1)

	_ = file.Close()

	if err != nil {
		return fmt.Errorf("ref/store/files: read dir %q: %w", dir, err)
	}

	for _, entry := range entries {
		name := path.Join(dir, entry.Name())
		if entry.IsDir() {
			err = collectLooseRefNames(root, name, names)
			if err != nil {
				return err
			}

			continue
		}

		if strings.HasSuffix(name, ".lock") {
			continue
		}

		names[name] = struct{}{}
	}

	return nil
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

// verifyPrepared re-reads every target under its lock
// and verifies the expected current values.
func (executor *updateExecutor) verifyPrepared(prepared []preparedUpdate) (string, error) {
	for i := range prepared {
		item := &prepared[i]

		refState, err := executor.directRead(item.target.name)
		if err != nil {
			return item.op.name, err
		}

		item.target.ref = refState

		err = verifyPreparedUpdateCurrent(*item)
		if err != nil {
			return item.op.name, err
		}
	}

	return "", nil
}

func verifyPreparedUpdateCurrent(item preparedUpdate) error {
	switch item.op.kind {
	case updateCreate, updateCreateSymbolic:
		if item.target.ref.kind != directMissing {
			return store.ErrCreateExists
		}

		return nil
	case updateReplace, updateDelete, updateVerify:
		if item.target.ref.kind == directMissing {
			return store.ErrReferenceNotFound
		}

		if item.target.ref.kind != directDirect {
			return store.ErrExpectedDirect
		}

		if item.target.ref.id != item.op.oldID {
			return &store.WrongOldIDError{Actual: item.target.ref.id, Expected: item.op.oldID}
		}

		return nil
	case updateReplaceSymbolic, updateDeleteSymbolic, updateVerifySymbolic:
		if item.target.ref.kind == directMissing {
			return store.ErrReferenceNotFound
		}

		if item.target.ref.kind != directSymbolic {
			return store.ErrExpectedSymbolic
		}

		if item.target.ref.target != item.op.oldTarget {
			return &store.WrongOldTargetError{Actual: item.target.ref.target, Expected: item.op.oldTarget}
		}

		return nil
	default:
		panic(fmt.Sprintf("ref/store/files: unsupported update operation %d", item.op.kind))
	}
}
