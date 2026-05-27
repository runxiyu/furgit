package packed

import (
	"os"

	objectstore "lindenii.org/go/furgit/object/store"
)

var _ objectstore.PackQuarantiner = (*Store)(nil)

type packQuarantine struct {
	*Store

	parent   *Store
	tempName string
	tempRoot *os.Root
}

var _ objectstore.PackQuarantine = (*packQuarantine)(nil)
