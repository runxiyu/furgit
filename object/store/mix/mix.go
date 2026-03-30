// Package mix provides an adaptive wrapper over multiple object storage
// backends.
package mix

import (
	"sync"

	objectstore "codeberg.org/lindenii/furgit/object/store"
)

// Mix queries multiple object databases with an MRU backend preference.
//
// Labels: Close-Caller.
type Mix struct {
	mu sync.RWMutex

	backendHead        *backendNode
	backendTail        *backendNode
	backendNodeByStore map[objectstore.Reader]*backendNode
}
