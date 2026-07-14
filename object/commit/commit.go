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
	Parents      []id.ObjectID //exhaustruct:optional
	Author       signature.Signature
	Committer    signature.Signature
	Message      []byte
	ChangeID     []byte        //exhaustruct:optional
	ExtraHeaders []ExtraHeader //exhaustruct:optional
}

// ExtraHeader represents an extra header in a Git object.
type ExtraHeader struct {
	Key   []byte
	Value []byte
}
