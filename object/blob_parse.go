package object

// ParseBlob decodes a blob object body.
func ParseBlob(body []byte) (*Blob, error) {
	return &Blob{Data: append([]byte(nil), body...)}, nil
}
