// Package files provides one Git files ref store with loose-over-packed reads
// and transaction-coordinated updates.
package files

import (
	"math/rand"
	"os"
	"time"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/refstore"
)

// Store reads and writes one Git files ref namespace rooted at one repository
// gitdir plus its commondir.
//
// Store borrows gitRoot and owns commonRoot. Close releases only resources
// opened by the store itself.
type Store struct {
	gitRoot    *os.Root
	commonRoot *os.Root
	algo       objectid.Algorithm
	lockRand   *rand.Rand

	packedRefsTimeout time.Duration
}

var (
	_ refstore.ReadingStore       = (*Store)(nil)
	_ refstore.TransactionalStore = (*Store)(nil)
	_ refstore.BatchStore         = (*Store)(nil)
)
