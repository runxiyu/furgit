package objectid

// Algorithm returns the object ID's hash algorithm.
func (id ObjectID) Algorithm() Algorithm {
	return id.algo
}
