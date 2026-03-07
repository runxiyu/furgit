package files

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/ref/refname"
	"codeberg.org/lindenii/furgit/refstore"
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
		err = tx.createPackedLock()
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

func (tx *Transaction) resolveOrdinaryTarget(name string, allowMissing bool) (resolvedWriteTarget, error) {
	cur := name
	seen := make(map[string]struct{})

	for {
		if _, ok := seen[cur]; ok {
			return resolvedWriteTarget{}, fmt.Errorf("refstore/files: symbolic reference cycle at %q", cur)
		}

		seen[cur] = struct{}{}

		refState, err := tx.directRead(cur)
		if err != nil {
			return resolvedWriteTarget{}, err
		}

		switch refState.kind {
		case directMissing:
			if !allowMissing {
				return resolvedWriteTarget{}, refstore.ErrReferenceNotFound
			}

			return resolvedWriteTarget{name: cur, loc: tx.store.loosePath(cur), ref: refState}, nil
		case directDetached:
			return resolvedWriteTarget{name: cur, loc: tx.store.loosePath(cur), ref: refState}, nil
		case directSymbolic:
			target := strings.TrimSpace(refState.target)
			if target == "" {
				return resolvedWriteTarget{}, fmt.Errorf("refstore/files: symbolic reference %q has empty target", cur)
			}

			cur = target
		default:
			return resolvedWriteTarget{}, fmt.Errorf("refstore/files: unsupported direct reference state %d", refState.kind)
		}
	}
}

func (tx *Transaction) directRead(name string) (directRef, error) {
	loc := tx.store.loosePath(name)
	hasPacked := false

	if loc.root == rootCommon && refname.ParseWorktree(name).Type == refname.WorktreeShared {
		packed, packedErr := tx.store.readPackedRefs()
		if packedErr != nil {
			return directRef{}, packedErr
		}

		_, hasPacked = packed.byName[name]
	}

	loose, err := tx.store.readLooseRef(name)
	if err == nil {
		switch loose := loose.(type) {
		case ref.Detached:
			return directRef{
				kind:     directDetached,
				name:     name,
				id:       loose.ID,
				isLoose:  true,
				isPacked: hasPacked,
			}, nil
		case ref.Symbolic:
			return directRef{
				kind:     directSymbolic,
				name:     name,
				target:   loose.Target,
				isLoose:  true,
				isPacked: hasPacked,
			}, nil
		default:
			return directRef{}, fmt.Errorf("refstore/files: unsupported reference type %T", loose)
		}
	}

	if !errors.Is(err, refstore.ErrReferenceNotFound) {
		info, statErr := tx.store.rootFor(loc.root).Stat(loc.path)
		if statErr != nil || !info.IsDir() {
			return directRef{}, err
		}
	}

	if hasPacked {
		packed, packedErr := tx.store.readPackedRefs()
		if packedErr != nil {
			return directRef{}, packedErr
		}

		detached := packed.byName[name]

		return directRef{
			kind:     directDetached,
			name:     name,
			id:       detached.ID,
			isPacked: true,
		}, nil
	}

	return directRef{
		kind: directMissing,
		name: name,
	}, nil
}

func (tx *Transaction) visibleNames() (map[string]struct{}, error) {
	names := make(map[string]struct{})

	looseNames, err := tx.store.collectLooseRefNames()
	if err != nil {
		return nil, err
	}

	for _, name := range looseNames {
		names[name] = struct{}{}
	}

	packed, err := tx.store.readPackedRefs()
	if err != nil {
		return nil, err
	}

	for name := range packed.byName {
		if _, exists := names[name]; exists {
			continue
		}

		names[name] = struct{}{}
	}

	return names, nil
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
			return fmt.Errorf("refstore/files: reference name conflict between %q and %q", name, existingName)
		}
	}

	for _, other := range writes {
		if other == name {
			continue
		}

		if refnamesConflict(name, other) {
			return fmt.Errorf("refstore/files: reference name conflict between %q and %q", name, other)
		}
	}

	return nil
}

func refnamesConflict(left, right string) bool {
	return left == right ||
		strings.HasPrefix(left, right+"/") ||
		strings.HasPrefix(right, left+"/")
}
