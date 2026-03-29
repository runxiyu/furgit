package objectid

// ParseSignatureHeaderName parses one canonical signature header name such as
// "gpgsig" or "gpgsig-sha256".
func ParseSignatureHeaderName(s string) (Algorithm, bool) {
	algo, ok := algorithmBySignatureHeaderName[s]

	return algo, ok
}
