package tree

import objecttype "codeberg.org/lindenii/furgit/object/type"

// ObjectType returns TypeTree.
func (tree *Tree) ObjectType() objecttype.Type {
	_ = tree

	return objecttype.TypeTree
}
