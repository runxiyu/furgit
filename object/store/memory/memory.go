package memory

import (
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/typ"
	"lindenii.org/go/lgo/sync"
)

// Memory is one in-memory object store.
//
// Labels: MT-Safe, Close-Caller.
type Memory struct {
	objectFormat id.ObjectFormat
	objects      *sync.Map[id.ObjectID, storedObject]
}

// storedObject is one in-memory object entry.
type storedObject struct {
	ty      typ.Type
	content []byte
}

var (
	_ store.ObjectReader = (*Memory)(nil)
	_ store.ObjectWriter = (*Memory)(nil)
)

// New builds one empty in-memory store for one object format.
func New(objectFormat id.ObjectFormat) *Memory {
	return &Memory{
		objectFormat: objectFormat,
		objects:      &sync.Map[id.ObjectID, storedObject]{},
	}
}

// ObjectFormat returns the object format used by the store.
func (memory *Memory) ObjectFormat() id.ObjectFormat {
	return memory.objectFormat
}
