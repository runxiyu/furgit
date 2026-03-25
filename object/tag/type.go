package tag

import objecttype "codeberg.org/lindenii/furgit/object/type"

// ObjectType returns TypeTag.
func (tag *Tag) ObjectType() objecttype.Type {
	_ = tag

	return objecttype.TypeTag
}
