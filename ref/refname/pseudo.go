package refname

// IsPseudo reports whether name is one Git pseudo-ref.
func IsPseudo(name string) bool {
	switch name {
	case "FETCH_HEAD", "MERGE_HEAD":
		return true
	default:
		return false
	}
}
