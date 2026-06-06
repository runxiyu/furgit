package signature

// Signature represents a Git signature (author/committer/tagger).
//
// Labels: MT-Unsafe.
type Signature struct {
	Name          []byte
	Email         []byte
	WhenUnix      int64
	OffsetMinutes int32
}
