package files

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

func (tx *Transaction) createLock(name refPath) error {
	root := tx.store.rootFor(name.root)
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

func (tx *Transaction) createPackedLock() error {
	const (
		initialBackoffMs     = 1
		backoffMaxMultiplier = 1000
	)

	timeout := tx.store.packedRefsTimeout
	deadline := time.Now().Add(timeout)
	multiplier := 1
	n := 1

	for {
		file, err := tx.store.commonRoot.OpenFile("packed-refs.lock", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err == nil {
			return file.Close()
		}

		if !errors.Is(err, os.ErrExist) {
			return err
		}

		if timeout == 0 || (timeout > 0 && time.Now().After(deadline)) {
			return err
		}

		backoffMs := multiplier * initialBackoffMs
		waitMs := (750 + tx.store.lockRand.Intn(500)) * backoffMs / 1000
		time.Sleep(time.Duration(waitMs) * time.Millisecond)

		multiplier += 2*n + 1
		if multiplier > backoffMaxMultiplier {
			multiplier = backoffMaxMultiplier
		} else {
			n++
		}
	}
}

func (tx *Transaction) targetKey(name refPath) string {
	return fmt.Sprintf("%d:%s", name.root, name.path)
}

func refPathFromKey(key string) refPath {
	rootValue, pathValue, ok := strings.Cut(key, ":")
	if !ok || rootValue == "" {
		return refPath{root: rootCommon, path: key}
	}

	if rootValue == "0" {
		return refPath{root: rootGit, path: pathValue}
	}

	return refPath{root: rootCommon, path: pathValue}
}
