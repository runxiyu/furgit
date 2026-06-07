package tag

import "lindenii.org/go/furgit/object/id"

// AppendSignature appends the signature for objectFormat to dst,
// and reports whether the tag carries a signature for objectFormat.
func (tag *Tag) AppendSignature(dst []byte, objectFormat id.ObjectFormat) ([]byte, bool) {
	signature, ok := tag.signatures[objectFormat]
	if !ok {
		return dst, false
	}

	for _, part := range signature {
		dst = append(dst, tag.body[part.start:part.end]...)
	}

	return dst, true
}
