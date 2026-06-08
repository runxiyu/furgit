package stored

import (
	"lindenii.org/go/furgit/object"
	"lindenii.org/go/furgit/object/id"
)

// Stored represents a stored object,
// i.e., an object along with its object ID.
type Stored[T object.Object] struct {
	id  id.ObjectID
	obj T
}

// New creates one stored object wrapper.
func New[T object.Object](id id.ObjectID, obj T) *Stored[T] {
	return &Stored[T]{id: id, obj: obj}
}

// Object returns the wrapped object as itself.
func (stored *Stored[T]) Object() T { //nolint:ireturn
	return stored.obj
}

// ID returns the object ID.
func (stored *Stored[T]) ID() id.ObjectID {
	return stored.id
}
