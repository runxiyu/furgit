package tag

import objecttype "lindenii.org/go/furgit/object/type"

// ObjectType returns TypeTag.
func (tag *Tag) ObjectType() objecttype.Type {
	_ = tag

	return objecttype.TypeTag
}
