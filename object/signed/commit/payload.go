package commit

// AppendPayload appends the commit verification payload to dst,
// omitting all embedded signature headers.
func (commit *Commit) AppendPayload(dst []byte) []byte {
	for _, part := range commit.payload {
		dst = append(dst, commit.body[part.start:part.end]...)
	}

	return dst
}
