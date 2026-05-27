package signedcommit

import objectid "lindenii.org/go/furgit/object/id"

// AppendSignature appends the unfolded signature for algo to dst.
func (commit *Commit) AppendSignature(dst []byte, algo objectid.Algorithm) ([]byte, bool) {
	signature, ok := commit.signatures[algo]
	if !ok {
		return dst, false
	}

	for _, part := range signature {
		dst = append(dst, commit.body[part.start:part.end]...)
	}

	return dst, true
}
