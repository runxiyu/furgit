package commit

import (
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/signature"
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

// ExtraHeader represents an extra header in a Git object.
type ExtraHeader struct {
	Key   string
	Value []byte
}
