// Package signature provides routines and representations that implement author/committer/tagger signatures.
package signature

// Signature represents a Git signature (author/committer/tagger).
type Signature struct {
	Name          []byte
	Email         []byte
	WhenUnix      int64
	OffsetMinutes int32
}
