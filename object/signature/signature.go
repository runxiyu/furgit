package signature

import "time"

// Signature represents a Git signature (author/committer/tagger).
//
// Labels: MT-Unsafe.
type Signature struct {
	Name          []byte
	Email         []byte
	WhenUnix      int64
	OffsetMinutes int32
}

// When returns a time.Time with the signature's timezone offset.
func (signature Signature) When() time.Time {
	loc := time.FixedZone("git", int(signature.OffsetMinutes)*60)

	return time.Unix(signature.WhenUnix, 0).In(loc)
}
