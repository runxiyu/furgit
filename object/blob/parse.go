package blob

// Parse decodes a blob object body.
//
// The returned blob aliases body:
// its Data shares the same backing array,
// so the blob inherits body's lifetime
// and must not be mutated unless body may be.
// Use [Blob.Clone] for an independent copy.
//
// Labels: Life-Parent, Mut-No.
func Parse(body []byte) (*Blob, error) {
	return &Blob{Data: body}, nil
}
