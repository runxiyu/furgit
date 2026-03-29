package objectid

// Sum computes an object ID from raw data using the selected algorithm.
func (algo Algorithm) Sum(data []byte) ObjectID {
	return algo.info().sum(data)
}
