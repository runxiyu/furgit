package tag

// AppendPayload appends the tag verification payload to dst,
// omitting all embedded signatures.
func (tag *Tag) AppendPayload(dst []byte) []byte {
	for _, part := range tag.payload {
		dst = append(dst, tag.body[part.start:part.end]...)
	}

	return dst
}
