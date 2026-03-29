package objectid

// Zero returns the all-zero object ID for this algorithm.
func (algo Algorithm) Zero() ObjectID {
	id, err := FromBytes(algo, make([]byte, algo.Size()))
	if err != nil {
		panic(err)
	}

	return id
}
