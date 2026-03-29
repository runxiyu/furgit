package objectid

// SignatureHeaderName returns the signature header name for this algorithm.
func (algo Algorithm) SignatureHeaderName() string {
	return algo.info().signatureHeaderName
}
