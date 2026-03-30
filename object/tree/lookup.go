package tree

// Entry looks up a tree entry by name.
//
// The returned pointer refers to storage within tree.Entries and must not be
// retained across InsertEntry or RemoveEntry calls.
func (tree *Tree) Entry(name []byte) *TreeEntry {
	if len(tree.Entries) == 0 {
		return nil
	}

	index, ok := tree.entryIndex(name)
	if !ok {
		return nil
	}

	return &tree.Entries[index]
}
