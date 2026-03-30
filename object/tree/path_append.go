package tree

// AppendPath appends path to dst as one slash-separated byte path.
func AppendPath(dst []byte, path [][]byte) []byte {
	for i := range path {
		if i > 0 {
			dst = append(dst, '/')
		}

		dst = append(dst, path[i]...)
	}

	return dst
}
