package memory

import (
	"sync"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref/store"
)

// Memory reads and writes one in-memory Git reference namespace.
//
// Labels: Close-Caller.
type Memory struct {
	mu           sync.RWMutex //exhaustruct:optional
	objectFormat id.ObjectFormat
	refs         map[string]storedRef
}

var (
	_ store.Reader        = (*Memory)(nil)
	_ store.Transactioner = (*Memory)(nil)
	_ store.Batcher       = (*Memory)(nil)
)

// New builds one empty in-memory reference store for one object format.
func New(objectFormat id.ObjectFormat) *Memory {
	return &Memory{
		objectFormat: objectFormat,
		refs:         make(map[string]storedRef),
	}
}

// ObjectFormat returns the object format used by the store.
func (memory *Memory) ObjectFormat() id.ObjectFormat {
	return memory.objectFormat
}

// Close closes the in-memory reference store.
//
// Labels: MT-Unsafe.
func (memory *Memory) Close() error {
	return nil
}
