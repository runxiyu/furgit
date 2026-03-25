// Package mix provides an adaptive wrapper over multiple object storage
// backends.
package mix

import (
	"sync"

	objectstorer "codeberg.org/lindenii/furgit/object/storer"
)

// Mix queries multiple object databases with an MRU backend preference.
//
// Mix borrows its backend stores.
type Mix struct {
	mu sync.RWMutex

	backendHead        *backendNode
	backendTail        *backendNode
	backendNodeByStore map[objectstorer.Store]*backendNode
}
