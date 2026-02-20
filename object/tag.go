package object

import "codeberg.org/lindenii/furgit/oid"

// Tag represents a Git annotated tag object.
type Tag struct {
	Target     oid.ObjectID
	TargetType Type
	Name       []byte
	Tagger     *Ident
	Message    []byte
}

// ObjectType returns TypeTag.
func (tag *Tag) ObjectType() Type {
	_ = tag
	return TypeTag
}
