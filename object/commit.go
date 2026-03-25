package object

import (
	"codeberg.org/lindenii/furgit/objectid"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// Commit represents a Git commit object.
type Commit struct {
	Tree         objectid.ObjectID
	Parents      []objectid.ObjectID
	Author       Signature
	Committer    Signature
	Message      []byte
	ChangeID     string
	ExtraHeaders []ExtraHeader
}

// ObjectType returns TypeCommit.
func (commit *Commit) ObjectType() objecttype.Type {
	_ = commit

	return objecttype.TypeCommit
}
