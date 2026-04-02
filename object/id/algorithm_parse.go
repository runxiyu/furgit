package id

// ParseAlgorithm parses a canonical algorithm name (e.g. "sha1", "sha256").
func ParseAlgorithm(s string) (Algorithm, bool) {
	algo, ok := algorithmByName[s]

	return algo, ok
}

// ParseSignatureHeaderName parses one canonical signature header name
// such as "gpgsig" or "gpgsig-sha256" to its respective algorithm.
func ParseSignatureHeaderName(s string) (Algorithm, bool) {
	algo, ok := algorithmBySignatureHeaderName[s]

	return algo, ok
}
