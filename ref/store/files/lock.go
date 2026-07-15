package files

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path"
	"slices"
	"strings"
	"time"
)

// createLockFile creates lockPath exclusively,
// retrying with quadratic backoff while it already exists.
// A zero timeout means one attempt;
// a negative timeout means retrying without bound.
func createLockFile(root *os.Root, lockPath string, timeout time.Duration) error {
	const (
		initialBackoffMs     = 1
		backoffMaxMultiplier = 1000
	)

	deadline := time.Now().Add(timeout)
	multiplier := 1
	n := 1

	for {
		file, err := root.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err == nil {
			closeErr := file.Close()
			if closeErr != nil {
				return fmt.Errorf("ref/store/files: %w", closeErr)
			}

			return nil
		}

		if !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("ref/store/files: %w", err)
		}

		if timeout == 0 || (timeout > 0 && time.Now().After(deadline)) {
			return fmt.Errorf("ref/store/files: %w", err)
		}

		backoffMs := multiplier * initialBackoffMs
		waitMs := (750 + rand.IntN(500)) * backoffMs / 1000 //nolint:gosec
		time.Sleep(time.Duration(waitMs) * time.Millisecond)

		multiplier += 2*n + 1
		if multiplier > backoffMaxMultiplier {
			multiplier = backoffMaxMultiplier
		} else {
			n++
		}
	}
}

// lockPrepared locks every prepared target in sorted order.
// On failure it returns the name of the offending operation
// alongside the error.
func (executor *updateExecutor) lockPrepared(prepared []preparedUpdate) (string, error) {
	locs := make([]refPath, 0, len(prepared))
	for _, item := range prepared {
		locs = append(locs, item.target.loc)
	}

	slices.SortFunc(locs, compareRefPath)

	for _, loc := range locs {
		err := executor.lockRef(loc)
		if err != nil {
			return opNameForLoc(prepared, loc), err
		}
	}

	return "", nil
}

// lockRef creates the directory for one reference and its lock file.
func (executor *updateExecutor) lockRef(loc refPath) error {
	root := executor.files.root(loc.kind)

	dir := path.Dir(loc.path)
	if dir != "." {
		err := root.MkdirAll(dir, 0o755)
		if err != nil {
			return fmt.Errorf("ref/store/files: %w", err)
		}
	}

	err := createLockFile(root, loc.path+".lock", executor.files.options.LooseLockTimeout)
	if err != nil {
		return err
	}

	executor.lockedRefs = append(executor.lockedRefs, loc)

	return nil
}

// lockPackedRefs takes the packed-refs lock.
//
// The lock is held for the rest of the prepare/commit cycle
// even when no packed-refs rewrite turns out to be needed,
// so that a concurrent repacking cannot resurrect deleted references.
func (executor *updateExecutor) lockPackedRefs() error {
	err := createLockFile(executor.files.commonRoot, "packed-refs.lock", executor.files.options.PackedLockTimeout)
	if err != nil {
		return err
	}

	executor.packedLocked = true

	return nil
}

// cleanup releases every lock still held by the executor,
// and then prunes the empty parent directories of deleted references,
// which their own lock files kept non-empty until now.
func (executor *updateExecutor) cleanup() {
	for _, loc := range executor.lockedRefs {
		_ = executor.files.root(loc.kind).Remove(loc.path + ".lock")
	}

	executor.lockedRefs = nil

	if executor.packedLocked {
		_ = executor.files.commonRoot.Remove("packed-refs.lock")
		executor.packedLocked = false
	}

	for _, name := range executor.deletedNames {
		executor.files.tryRemoveEmptyParents(name, true, false)
	}

	executor.deletedNames = nil
}

// tryRemoveEmptyParents removes empty parent directories
// of one deleted reference name,
// in the loose reference tree, the reflog tree, or both.
// The first two name components are always preserved.
func (files *Files) tryRemoveEmptyParents(name string, pruneRefs, pruneLogs bool) {
	for _, prefix := range parentPrefixes(name) {
		loc := files.loosePath(prefix)
		root := files.root(loc.kind)

		if pruneRefs {
			err := root.Remove(loc.path)
			if err != nil {
				pruneRefs = false
			}
		}

		if pruneLogs {
			err := root.Remove("logs/" + loc.path)
			if err != nil {
				pruneLogs = false
			}
		}

		if !pruneRefs && !pruneLogs {
			return
		}
	}
}

// parentPrefixes returns the proper prefixes of name
// that are candidates for empty-parent removal,
// longest first,
// preserving the first two name components.
func parentPrefixes(name string) []string {
	var prefixes []string

	end := len(name)

	for {
		slash := strings.LastIndex(name[:end], "/")
		if slash < 0 {
			break
		}

		end = slash

		prefix := name[:end]
		if strings.Count(prefix, "/") < 2 {
			break
		}

		prefixes = append(prefixes, prefix)
	}

	return prefixes
}

func compareRefPath(left, right refPath) int {
	if left.kind != right.kind {
		return int(left.kind) - int(right.kind)
	}

	return strings.Compare(left.path, right.path)
}

func opNameForLoc(prepared []preparedUpdate, loc refPath) string {
	for _, item := range prepared {
		if item.target.loc == loc {
			return item.op.name
		}
	}

	return ""
}
