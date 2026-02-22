package object

import (
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// Tag represents a Git annotated tag object.
type Tag struct {
	Target     objectid.ObjectID
	TargetType objecttype.Type
	Name       []byte
	Tagger     *Signature
	Message    []byte
}

// ObjectType returns TypeTag.
func (tag *Tag) ObjectType() objecttype.Type {
	_ = tag
	return objecttype.TypeTag
}
