// Package blob provides representations, parsers, and serializers for blob objects.
package blob

// Blob represents a Git blob object.
//
// This Blob object is fully materialized in memory.
// Consider using objectstore/Store.ReadReaderContent,
// or appropriate streaming write APIs.
type Blob struct {
	Data []byte
}
