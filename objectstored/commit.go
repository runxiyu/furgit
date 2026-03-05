package objectstored

import (
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
)

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
