package loose

import (
	"os"

	objectstore "codeberg.org/lindenii/furgit/object/store"
)

var _ objectstore.ObjectQuarantiner = (*Store)(nil)

type objectQuarantine struct {
	*Store

	parent   *Store
	tempName string
	tempRoot *os.Root
}

var _ objectstore.ObjectQuarantine = (*objectQuarantine)(nil)
