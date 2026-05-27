package tree

import objecttype "lindenii.org/go/furgit/object/type"

// ObjectType returns TypeTree.
func (tree *Tree) ObjectType() objecttype.Type {
	_ = tree

	return objecttype.TypeTree
}
