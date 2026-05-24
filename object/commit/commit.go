package commit

import (
	"codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/signature"
)

// Commit represents a fully materialized Git commit object.
//
// Labels: MT-Unsafe.
type Commit struct {
	Tree         id.ObjectID
	Parents      []id.ObjectID
	Author       signature.Signature
	Committer    signature.Signature
	Message      []byte
	ChangeID     string
	ExtraHeaders []ExtraHeader
}
