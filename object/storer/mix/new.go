package mix

import objectstorer "codeberg.org/lindenii/furgit/object/storer"

// New creates a Mix from backends.
//
// The provided backends must be non-nil and distinct.
// Mix borrows the provided backends and does not close them in Close.
func New(backends ...objectstorer.Store) *Mix {
	nodeByStore := make(map[objectstorer.Store]*backendNode, len(backends))

	var (
		head *backendNode
		tail *backendNode
	)

	for _, backend := range backends {
		node := &backendNode{
			backend: backend,
			prev:    tail,
		}
		if tail != nil {
			tail.next = node
		}

		if head == nil {
			head = node
		}

		tail = node
		nodeByStore[backend] = node
	}

	return &Mix{
		backendHead:        head,
		backendTail:        tail,
		backendNodeByStore: nodeByStore,
	}
}
