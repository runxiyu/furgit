// Package commit provides parsed commit objects and commit serialization.
//
// It parses commits into ordinary Go values for reading and construction. It
// does not preserve the exact original byte layout needed for signature
// verification; callers that need signature-verification payload fidelity
// should use [lindenii.org/go/furgit/object/signed/commit].
package commit

import (
	objectid "lindenii.org/go/furgit/object/id"
	objectsignature "lindenii.org/go/furgit/object/signature"
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
