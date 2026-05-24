package blob

// Parse decodes a blob object body.
//
// Labels: Deps-Owned, Life-Independent
func Parse(body []byte) (*Blob, error) {
	return &Blob{Data: append([]byte(nil), body...)}, nil
}
