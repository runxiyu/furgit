package objectid

// EmptyTree returns the object ID of an empty tree ("tree 0\x00") for this
// algorithm.
func (algo Algorithm) EmptyTree() ObjectID {
	return algo.info().emptyTree
}
