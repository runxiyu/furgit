package commit

import "lindenii.org/go/furgit/object/id"

// AppendSignature appends the unfolded signature for objectFormat to dst,
// and reports whether the commit carries a signature for objectFormat.
func (commit *Commit) AppendSignature(dst []byte, objectFormat id.ObjectFormat) ([]byte, bool) {
	signature, ok := commit.signatures[objectFormat]
	if !ok {
		return dst, false
	}

	for _, part := range signature {
		dst = append(dst, commit.body[part.start:part.end]...)
	}

	return dst, true
}
