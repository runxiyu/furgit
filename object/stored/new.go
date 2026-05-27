package stored

import (
	"lindenii.org/go/furgit/object"
	objectid "lindenii.org/go/furgit/object/id"
)

// New creates one stored object wrapper.
func New[T object.Object](id objectid.ObjectID, obj T) *Stored[T] {
	return &Stored[T]{id: id, obj: obj}
}
