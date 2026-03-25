package files

import (
	"os"
	"path"
)

func (executor *refUpdateExecutor) createUpdateLock(name refPath) error {
	root := executor.store.rootFor(name.root)
	dir := path.Dir(name.path)

	if dir != "." {
		err := root.MkdirAll(dir, 0o755)
		if err != nil {
			return err
		}
	}

	file, err := root.OpenFile(name.path+".lock", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}

	return file.Close()
}
