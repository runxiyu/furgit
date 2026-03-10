package ingest

import "io"

// countingWriter counts bytes written to dst.
type countingWriter struct {
	dst io.Writer
	n   int
}

// Write writes src to dst and tracks output byte count.
func (writer *countingWriter) Write(src []byte) (int, error) {
	n, err := writer.dst.Write(src)
	writer.n += n

	return n, err
}
