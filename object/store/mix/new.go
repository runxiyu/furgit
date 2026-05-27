package mix

import objectstore "lindenii.org/go/furgit/object/store"

// New creates a Mix from backends.
//
// The provided backends must be non-nil and distinct.
//
// Labels: Deps-Borrowed, Life-Parent.
func New(backends ...objectstore.Reader) *Mix {
	nodeByStore := make(map[objectstore.Reader]*backendNode, len(backends))

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
