package files

import (
	"os"
)

func (store *Store) rootFor(kind rootKind) *os.Root {
	if kind == rootCommon {
		return store.commonRoot
	}

	return store.gitRoot
}
