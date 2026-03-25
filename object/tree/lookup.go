package tree

// Entry looks up a tree entry by name.
func (tree *Tree) Entry(name []byte) *TreeEntry {
	if len(tree.Entries) == 0 {
		return nil
	}

	if e := tree.entry(name, true); e != nil {
		return e
	}

	return tree.entry(name, false)
}
