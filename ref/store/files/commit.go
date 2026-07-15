package files

import (
	"errors"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"lindenii.org/go/furgit/ref"
)

// errNonEmptyDirectory indicates that a non-empty directory
// occupies the place of a reference about to be written.
var errNonEmptyDirectory = errors.New("ref/store/files: non-empty directory blocks reference")

// commit applies one prepared operation set,
// writing loose updates first,
// then removing reflogs and packed and loose versions of deleted references,
// and finally releasing all locks.
func (executor *updateExecutor) commit(prepared []preparedUpdate) error {
	defer executor.cleanup()

	for _, item := range prepared {
		switch item.op.kind {
		case updateCreate, updateReplace, updateCreateSymbolic, updateReplaceSymbolic:
			err := executor.writeLooseUpdate(item)
			if err != nil {
				return err
			}
		case updateDelete, updateVerify, updateDeleteSymbolic, updateVerifySymbolic:
		default:
			panic(fmt.Sprintf("ref/store/files: unsupported update operation %d", item.op.kind))
		}
	}

	executor.removeDeletedReflogs(prepared)

	err := executor.applyPackedDeletes(prepared)
	if err != nil {
		return err
	}

	return executor.removeDeletedLooseRefs(prepared)
}

// writeLooseUpdate writes one new reference value into its lock file
// and renames the lock file into place.
func (executor *updateExecutor) writeLooseUpdate(item preparedUpdate) error {
	root := executor.files.root(item.target.loc.kind)
	lockName := item.target.loc.path + ".lock"

	var content string

	switch item.op.kind {
	case updateCreate, updateReplace:
		content = item.op.newID.String() + "\n"
	case updateCreateSymbolic, updateReplaceSymbolic:
		content = "ref: " + item.op.newTarget + "\n"
	case updateDelete, updateVerify, updateDeleteSymbolic, updateVerifySymbolic:
		panic(fmt.Sprintf("ref/store/files: unsupported write operation %d", item.op.kind))
	default:
		panic(fmt.Sprintf("ref/store/files: unsupported write operation %d", item.op.kind))
	}

	err := executor.writeIntoLock(root, lockName, content)
	if err != nil {
		return fmt.Errorf("ref/store/files: update %q: %w", item.op.name, err)
	}

	err = executor.removeEmptyDirTree(item.target.loc)
	if err != nil {
		return fmt.Errorf("ref/store/files: update %q: %w", item.op.name, err)
	}

	err = root.Rename(lockName, item.target.loc.path)
	if err != nil {
		return fmt.Errorf("ref/store/files: update %q: %w", item.op.name, err)
	}

	return nil
}

// writeIntoLock writes content into one already-created lock file,
// flushing it when the store is configured to.
func (executor *updateExecutor) writeIntoLock(root *os.Root, lockName, content string) error {
	lock, err := root.OpenFile(lockName, os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err //nolint:wrapcheck
	}

	_, err = lock.WriteString(content)
	if err != nil {
		_ = lock.Close()

		return err //nolint:wrapcheck
	}

	if executor.files.options.Fsync == FsyncAlways {
		err = lock.Sync()
		if err != nil {
			_ = lock.Close()

			return err //nolint:wrapcheck
		}
	}

	return lock.Close() //nolint:wrapcheck
}

// removeEmptyDirTree removes one directory tree of empty directories
// occupying the place of a reference about to be written.
func (executor *updateExecutor) removeEmptyDirTree(loc refPath) error {
	root := executor.files.root(loc.kind)

	info, err := root.Stat(loc.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err //nolint:wrapcheck
	}

	if !info.IsDir() {
		return nil
	}

	return executor.removeEmptyDirTreeRecursive(root, loc.path)
}

func (executor *updateExecutor) removeEmptyDirTreeRecursive(root *os.Root, dirPath string) error {
	dir, err := root.Open(dirPath)
	if err != nil {
		return err //nolint:wrapcheck
	}

	entries, err := dir.ReadDir(-1)

	_ = dir.Close()

	if err != nil {
		return err //nolint:wrapcheck
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			return fmt.Errorf("%w: %q", errNonEmptyDirectory, dirPath)
		}

		err = executor.removeEmptyDirTreeRecursive(root, path.Join(dirPath, entry.Name()))
		if err != nil {
			return err
		}
	}

	return root.Remove(dirPath) //nolint:wrapcheck
}

// removeDeletedReflogs removes the reflog files of deleted references
// before the references themselves are deleted.
func (executor *updateExecutor) removeDeletedReflogs(prepared []preparedUpdate) {
	for _, item := range prepared {
		switch item.op.kind {
		case updateDelete, updateDeleteSymbolic:
			root := executor.files.root(item.target.loc.kind)

			err := root.Remove("logs/" + item.target.loc.path)
			if err == nil {
				executor.files.tryRemoveEmptyParents(item.target.name, false, true)
			}
		case updateCreate, updateReplace, updateVerify,
			updateCreateSymbolic, updateReplaceSymbolic, updateVerifySymbolic:
		default:
			panic(fmt.Sprintf("ref/store/files: unsupported update operation %d", item.op.kind))
		}
	}
}

// applyPackedDeletes rewrites packed-refs without the deleted references,
// keeping the packed-refs lock when no rewrite is needed.
func (executor *updateExecutor) applyPackedDeletes(prepared []preparedUpdate) error {
	if !executor.packedLocked {
		return nil
	}

	deleted := make(map[string]struct{})
	needed := false

	for _, item := range prepared {
		if item.op.kind != updateDelete && item.op.kind != updateDeleteSymbolic {
			continue
		}

		deleted[item.target.name] = struct{}{}

		if item.target.ref.isPacked {
			needed = true
		}
	}

	if !needed {
		return nil
	}

	packed, err := executor.files.readPackedRefs()
	if err != nil {
		return err
	}

	survivors := make([]ref.Direct, 0, len(packed.ordered))

	for _, entry := range packed.ordered {
		if _, skip := deleted[entry.RefName]; skip {
			continue
		}

		survivors = append(survivors, entry)
	}

	slices.SortFunc(survivors, func(left, right ref.Direct) int {
		return strings.Compare(left.RefName, right.RefName)
	})

	err = executor.writeIntoLock(
		executor.files.commonRoot,
		"packed-refs.lock",
		formatPackedRefs(packed.traits, survivors),
	)
	if err != nil {
		return fmt.Errorf("ref/store/files: rewrite packed-refs: %w", err)
	}

	err = executor.files.commonRoot.Rename("packed-refs.lock", "packed-refs")
	if err != nil {
		return fmt.Errorf("ref/store/files: rewrite packed-refs: %w", err)
	}

	executor.packedLocked = false

	return nil
}

// formatPackedRefs serializes packed-refs content,
// propagating the input file's peel traits.
//
// TODO: Possibility of peeling objects here.
func formatPackedRefs(traits packedTraits, entries []ref.Direct) string {
	headerTraits := make([]string, 0, 3)

	if traits.peeled {
		headerTraits = append(headerTraits, "peeled")
	}

	if traits.fullyPeeled {
		headerTraits = append(headerTraits, "fully-peeled")
	}

	headerTraits = append(headerTraits, "sorted")

	var builder strings.Builder

	builder.WriteString(packedRefsHeaderPrefix)
	builder.WriteString(strings.Join(headerTraits, " "))
	builder.WriteString("\n")

	for _, entry := range entries {
		builder.WriteString(entry.ID.String())
		builder.WriteString(" ")
		builder.WriteString(entry.RefName)
		builder.WriteString("\n")

		if entry.PeelState == ref.PeelTo {
			builder.WriteString("^")
			builder.WriteString(entry.PeeledID.String())
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// removeDeletedLooseRefs removes the loose files of deleted references.
func (executor *updateExecutor) removeDeletedLooseRefs(prepared []preparedUpdate) error {
	for _, item := range prepared {
		switch item.op.kind {
		case updateDelete, updateDeleteSymbolic:
			if item.target.ref.isLoose {
				root := executor.files.root(item.target.loc.kind)

				err := root.Remove(item.target.loc.path)
				if err != nil && !errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("ref/store/files: delete %q: %w", item.op.name, err)
				}
			}

			executor.deletedNames = append(executor.deletedNames, item.target.name)
		case updateCreate, updateReplace, updateVerify,
			updateCreateSymbolic, updateReplaceSymbolic, updateVerifySymbolic:
		default:
			panic(fmt.Sprintf("ref/store/files: unsupported update operation %d", item.op.kind))
		}
	}

	return nil
}
