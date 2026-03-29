package signedtag

import objectid "codeberg.org/lindenii/furgit/object/id"

// AppendSignature appends the signature for algo to dst.
func (tag *Tag) AppendSignature(dst []byte, algo objectid.Algorithm) ([]byte, bool) {
	signature, ok := tag.signatures[algo]
	if !ok {
		return dst, false
	}

	for _, part := range signature {
		dst = append(dst, tag.body[part.start:part.end]...)
	}

	return dst, true
}
