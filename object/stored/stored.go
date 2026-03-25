// Package stored wraps parsed objects with their storage object IDs.
//
// Stored values are typically instantiated with pointer object types such as
// *blob.Blob, *tree.Tree, *commit.Commit, or *tag.Tag, because those
// pointer types satisfy object.Object.
package stored

import (
	"codeberg.org/lindenii/furgit/object"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

// Stored represents a stored object,
// i.e., an object along with its object ID.
type Stored[T object.Object] struct {
	id  objectid.ObjectID
	obj T
}

// New creates one stored object wrapper.
func New[T object.Object](id objectid.ObjectID, obj T) *Stored[T] {
	return &Stored[T]{id: id, obj: obj}
}

// ID returns the object ID.
func (stored *Stored[T]) ID() objectid.ObjectID {
	return stored.id
}

// Object returns the wrapped object as itself.
func (stored *Stored[T]) Object() T {
	return stored.obj
}
