// Package files provides one Git files ref store with loose-over-packed reads
// and transaction-coordinated updates.
package files

import (
	"math/rand"
	"os"
	"time"

	objectid "codeberg.org/lindenii/furgit/object/id"
	refstore "codeberg.org/lindenii/furgit/ref/store"
)

// Store reads and writes one Git files ref namespace rooted at one repository
// gitdir plus its commondir.
//
// Labels: Close-Caller.
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
