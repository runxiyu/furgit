package packed

import "io"

// readCloser proxies reads and closes one underlying closer.
type readCloser struct {
	reader io.Reader
	closer io.Closer
}

// Read proxies reads to the underlying reader.
func (reader *readCloser) Read(dst []byte) (int, error) {
	return reader.reader.Read(dst)
}

// Close closes the underlying closer.
func (reader *readCloser) Close() error {
	return reader.closer.Close()
}
