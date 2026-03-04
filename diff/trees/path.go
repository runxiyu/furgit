package trees

func joinPath(prefix, name []byte) []byte {
	if len(prefix) == 0 {
		out := make([]byte, len(name))
		copy(out, name)

		return out
	}

	out := make([]byte, len(prefix)+1+len(name))
	copy(out, prefix)
	out[len(prefix)] = '/'
	copy(out[len(prefix)+1:], name)

	return out
}
