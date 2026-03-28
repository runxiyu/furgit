package mix

import objectstore "codeberg.org/lindenii/furgit/object/store"

// New creates a Mix from backends.
//
// The provided backends must be non-nil and distinct.
// Mix borrows the provided backends and does not close them in Close.
func New(backends ...objectstore.ReadingStore) *Mix {
	nodeByStore := make(map[objectstore.ReadingStore]*backendNode, len(backends))

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
