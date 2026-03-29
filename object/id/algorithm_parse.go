package objectid

// ParseAlgorithm parses a canonical algorithm name (e.g. "sha1", "sha256").
func ParseAlgorithm(s string) (Algorithm, bool) {
	algo, ok := algorithmByName[s]

	return algo, ok
}
