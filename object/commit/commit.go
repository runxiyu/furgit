// Package commit provides representations, parsers, and serializers for commit objects.
package commit

import (
	objectid "codeberg.org/lindenii/furgit/object/id"
	objectsignature "codeberg.org/lindenii/furgit/object/signature"
)

// Commit represents a fully materialized Git commit object.
//
// Labels: MT-Unsafe.
type Commit struct {
	Tree         objectid.ObjectID
	Parents      []objectid.ObjectID
	Author       objectsignature.Signature
	Committer    objectsignature.Signature
	Message      []byte
	ChangeID     string
	ExtraHeaders []ExtraHeader
}
