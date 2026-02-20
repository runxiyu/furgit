package object

import "codeberg.org/lindenii/furgit/oid"

// Commit represents a Git commit object.
type Commit struct {
	Tree         oid.ObjectID
	Parents      []oid.ObjectID
	Author       Ident
	Committer    Ident
	Message      []byte
	ChangeID     string
	ExtraHeaders []ExtraHeader
}

// ObjectType returns TypeCommit.
func (commit *Commit) ObjectType() Type {
	_ = commit
	return TypeCommit
}
