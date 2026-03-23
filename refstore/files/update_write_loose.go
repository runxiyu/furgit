package files

import (
	"fmt"
	"os"
	"path"
	"strings"
)

func (executor *refUpdateExecutor) writePreparedLooseUpdate(item preparedUpdate) error {
	root := executor.store.rootFor(item.target.loc.root)
	lockName := item.target.loc.path + ".lock"

	lock, err := root.OpenFile(lockName, os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	var content string

	switch item.op.kind {
	case updateCreate, updateReplace:
		content = item.op.newID.String() + "\n"
	case updateCreateSymbolic, updateReplaceSymbolic:
		content = "ref: " + strings.TrimSpace(item.op.newTarget) + "\n"
	case updateDelete, updateVerify, updateDeleteSymbolic, updateVerifySymbolic:
	default:
		_ = lock.Close()

		return fmt.Errorf("refstore/files: unsupported write operation %d", item.op.kind)
	}

	_, err = lock.WriteString(content)
	if err != nil {
		_ = lock.Close()

		return err
	}

	err = lock.Close()
	if err != nil {
		return err
	}

	dir := path.Dir(item.target.loc.path)
	if dir != "." {
		err = root.MkdirAll(dir, 0o755)
		if err != nil {
			return err
		}
	}

	err = executor.removeEmptyDirTree(item.target.loc)
	if err != nil {
		return err
	}

	return root.Rename(lockName, item.target.loc.path)
}
