package object

import (
	"codeberg.org/lindenii/furgit/objecttype"
	"codeberg.org/lindenii/furgit/objectid"
)

// Tag represents a Git annotated tag object.
type Tag struct {
	Target     objectid.ObjectID
	TargetType objecttype.Type
	Name       []byte
	Tagger     *Ident
	Message    []byte
}

// ObjectType returns TypeTag.
func (tag *Tag) ObjectType() objecttype.Type {
	_ = tag
	return objecttype.TypeTag
}
