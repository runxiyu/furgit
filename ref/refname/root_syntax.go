package refname

// IsRootSyntax reports whether name matches Git's all-caps root-ref syntax.
func IsRootSyntax(name string) bool {
	for i := range len(name) {
		ch := name[i]
		if (ch < 'A' || ch > 'Z') && ch != '-' && ch != '_' {
			return false
		}
	}

	return true
}
