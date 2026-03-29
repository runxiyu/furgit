package objectid

// Zero returns the all-zero object ID for the specified algorithm.
func Zero(algo Algorithm) ObjectID {
	id, err := FromBytes(algo, make([]byte, algo.Size()))
	if err != nil {
		panic(err)
	}

	return id
}
