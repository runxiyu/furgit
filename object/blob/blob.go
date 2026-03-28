// Package blob provides representations, parsers, and serializers for blob objects.
package blob

// Blob represents a Git blob object.
//
// Blob is fully materialized in memory.
//
// Consider using objectstore.ReadingStore.ReadReaderContent,
// or appropriate streaming write APIs.
//
// Labels: MT-Unsafe.
type Blob struct {
	Data []byte
}
