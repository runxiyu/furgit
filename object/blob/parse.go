package blob

// Parse decodes a blob object body.
func Parse(body []byte) (*Blob, error) {
	return &Blob{Data: append([]byte(nil), body...)}, nil
}
