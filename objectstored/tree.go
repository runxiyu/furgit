package objectstored

import (
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
)

// StoredTree is a parsed tree paired with its storage ID.
type StoredTree struct {
	id   objectid.ObjectID
	tree *object.Tree
}

// NewStoredTree creates one stored tree wrapper.
func NewStoredTree(id objectid.ObjectID, tree *object.Tree) *StoredTree {
	return &StoredTree{id: id, tree: tree}
}

// ID returns the object ID this tree was loaded from.
func (stored *StoredTree) ID() objectid.ObjectID {
	return stored.id
}

// Object returns the parsed tree as the generic object interface.
func (stored *StoredTree) Object() object.Object {
	return stored.tree
}

// Tree returns the parsed tree value.
func (stored *StoredTree) Tree() *object.Tree {
	return stored.tree
}
