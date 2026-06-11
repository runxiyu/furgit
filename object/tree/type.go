package tree

import "lindenii.org/go/furgit/object/typ"

// ObjectType returns TypeTree.
func (tree *Tree) ObjectType() typ.Type {
	_ = tree

	return typ.Tree
}
