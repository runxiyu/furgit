package tag

import "lindenii.org/go/furgit/object/typ"

// ObjectType returns TypeTag.
func (tag *Tag) ObjectType() typ.Type {
	_ = tag

	return typ.Tag
}
