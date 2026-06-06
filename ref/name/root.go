package name

import "strings"

// IsPseudo reports whether name is a Git pseudo-ref.
func IsPseudo(name string) bool {
	switch name {
	case "FETCH_HEAD", "MERGE_HEAD":
		return true
	default:
		return false
	}
}

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

// IsRoot reports whether name is a root ref.
func IsRoot(name string) bool {
	if !IsRootSyntax(name) || IsPseudo(name) {
		return false
	}

	if strings.HasSuffix(name, "_HEAD") {
		return true
	}

	switch name {
	case "HEAD", "AUTO_MERGE", "BISECT_EXPECTED_REV", "NOTES_MERGE_PARTIAL", "NOTES_MERGE_REF", "MERGE_AUTOSTASH":
		return true
	default:
		return false
	}
}
