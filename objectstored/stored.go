// Package objectstored wraps parsed objects with their storage object IDs.
package objectstored

import (
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
)

// StoredObject is a parsed object paired with its storage ID.
type StoredObject interface {
	// ID returns the object ID the object was loaded from.
	ID() objectid.ObjectID
	// Object returns the parsed object value.
	Object() object.Object
}

// StoredBlob is a parsed blob paired with its storage ID.
type StoredBlob struct {
	id   objectid.ObjectID
	blob *object.Blob
}

// NewStoredBlob creates one stored blob wrapper.
func NewStoredBlob(id objectid.ObjectID, blob *object.Blob) *StoredBlob {
	return &StoredBlob{id: id, blob: blob}
}

// ID returns the object ID this blob was loaded from.
func (stored *StoredBlob) ID() objectid.ObjectID {
	return stored.id
}

// Object returns the parsed blob as the generic object interface.
func (stored *StoredBlob) Object() object.Object {
	return stored.blob
}

// Blob returns the parsed blob value.
func (stored *StoredBlob) Blob() *object.Blob {
	return stored.blob
}

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

// StoredCommit is a parsed commit paired with its storage ID.
type StoredCommit struct {
	id     objectid.ObjectID
	commit *object.Commit
}

// NewStoredCommit creates one stored commit wrapper.
func NewStoredCommit(id objectid.ObjectID, commit *object.Commit) *StoredCommit {
	return &StoredCommit{id: id, commit: commit}
}

// ID returns the object ID this commit was loaded from.
func (stored *StoredCommit) ID() objectid.ObjectID {
	return stored.id
}

// Object returns the parsed commit as the generic object interface.
func (stored *StoredCommit) Object() object.Object {
	return stored.commit
}

// Commit returns the parsed commit value.
func (stored *StoredCommit) Commit() *object.Commit {
	return stored.commit
}

// StoredTag is a parsed tag paired with its storage ID.
type StoredTag struct {
	id  objectid.ObjectID
	tag *object.Tag
}

// NewStoredTag creates one stored tag wrapper.
func NewStoredTag(id objectid.ObjectID, tag *object.Tag) *StoredTag {
	return &StoredTag{id: id, tag: tag}
}

// ID returns the object ID this tag was loaded from.
func (stored *StoredTag) ID() objectid.ObjectID {
	return stored.id
}

// Object returns the parsed tag as the generic object interface.
func (stored *StoredTag) Object() object.Object {
	return stored.tag
}

// Tag returns the parsed tag value.
func (stored *StoredTag) Tag() *object.Tag {
	return stored.tag
}
