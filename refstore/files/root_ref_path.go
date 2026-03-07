package files

import (
	"fmt"
	"strings"
)

type refPath struct {
	root rootKind
	path string
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
