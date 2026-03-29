package objectid

// String returns the canonical algorithm name.
func (algo Algorithm) String() string {
	inf := algo.info()
	if inf.name == "" {
		return "unknown"
	}

	return inf.name
}
