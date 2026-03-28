package mix

import objectstore "codeberg.org/lindenii/furgit/object/store"

type backendNode struct {
	backend objectstore.ReadingStore
	prev    *backendNode
	next    *backendNode
}

//nolint:ireturn
func (mix *Mix) firstBackend() objectstore.ReadingStore {
	mix.mu.RLock()
	defer mix.mu.RUnlock()

	if mix.backendHead == nil {
		return nil
	}

	return mix.backendHead.backend
}

//nolint:ireturn
func (mix *Mix) nextBackend(current objectstore.ReadingStore) objectstore.ReadingStore {
	mix.mu.RLock()
	defer mix.mu.RUnlock()

	node := mix.backendNodeByStore[current]
	if node == nil || node.next == nil {
		return nil
	}

	return node.next.backend
}

func (mix *Mix) touchBackend(backend objectstore.ReadingStore) {
	if backend == nil {
		return
	}

	if !mix.mu.TryLock() {
		return
	}
	defer mix.mu.Unlock()

	node := mix.backendNodeByStore[backend]
	if node == nil || node == mix.backendHead {
		return
	}

	if node.prev != nil {
		node.prev.next = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	}

	if mix.backendTail == node {
		mix.backendTail = node.prev
	}

	node.prev = nil

	node.next = mix.backendHead
	if mix.backendHead != nil {
		mix.backendHead.prev = node
	}

	mix.backendHead = node
	if mix.backendTail == nil {
		mix.backendTail = node
	}
}
