// Package tag provides parsed annotated tag objects and tag serialization.
//
// It parses annotated tags into ordinary Go values for reading and
// construction. It does not preserve the exact original byte layout needed for
// signature verification; callers that need signature-verification payload
// fidelity should use [lindenii.org/go/furgit/object/signed/tag].
package tag

import (
	objectid "lindenii.org/go/furgit/object/id"
	objectsignature "lindenii.org/go/furgit/object/signature"
	objecttype "lindenii.org/go/furgit/object/type"
)

// Tag represents a fully materialized Git annotated tag object.
//
// Labels: MT-Unsafe.
type Tag struct {
	TargetID   objectid.ObjectID
	TargetType objecttype.Type
	Name       []byte
	Tagger     *objectsignature.Signature
	Message    []byte
}
