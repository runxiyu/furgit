package stored

import (
	"codeberg.org/lindenii/furgit/object"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

// New creates one stored object wrapper.
func New[T object.Object](id objectid.ObjectID, obj T) *Stored[T] {
	return &Stored[T]{id: id, obj: obj}
}
