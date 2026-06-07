package tag

import (
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/signature"
	"lindenii.org/go/furgit/object/typ"
)

// Tag represents a fully materialized Git annotated tag object.
//
// Labels: MT-Unsafe.
type Tag struct {
	TargetID     id.ObjectID
	TargetType   typ.Type
	Name         []byte
	Tagger       signature.Signature
	Message      []byte
	ExtraHeaders []ExtraHeader
}

// ExtraHeader represents an extra header in a Git tag object.
type ExtraHeader struct {
	Key   string
	Value []byte
}
