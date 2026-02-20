package object

import (
	"codeberg.org/lindenii/furgit/objecttype"
	"codeberg.org/lindenii/furgit/objectid"
)

// Commit represents a Git commit object.
type Commit struct {
	Tree         objectid.ObjectID
	Parents      []objectid.ObjectID
	Author       Ident
	Committer    Ident
	Message      []byte
	ChangeID     string
	ExtraHeaders []ExtraHeader
}

// ObjectType returns TypeCommit.
func (commit *Commit) ObjectType() objecttype.Type {
	_ = commit
	return objecttype.TypeCommit
}
