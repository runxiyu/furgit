// Package mix provides an adaptive wrapper over multiple object storage
// backends.
package mix

import (
	"sync"

	"codeberg.org/lindenii/furgit/objectstore"
)

// Mix queries multiple object databases with an MRU backend preference.
type Mix struct {
	mu sync.RWMutex

	backendHead        *backendNode
	backendTail        *backendNode
	backendNodeByStore map[objectstore.Store]*backendNode
}
