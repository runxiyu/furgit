package signature

import "time"

// When returns a time.Time with the signature's timezone offset.
func (signature Signature) When() time.Time {
	loc := time.FixedZone("git", int(signature.OffsetMinutes)*60)

	return time.Unix(signature.WhenUnix, 0).In(loc)
}
