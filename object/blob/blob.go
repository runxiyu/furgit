package blob

// Blob represents a Git blob object.
//
// Blob is fully materialized in memory.
// Consider using [lindenii.org/go/furgit/object/fetch]
// for streaming access, or appropriate streaming write APIs.
//
// Labels: MT-Unsafe.
type Blob struct {
	Data []byte
}
