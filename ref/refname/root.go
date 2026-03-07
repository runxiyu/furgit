package refname

import "strings"

// IsRoot reports whether name is one root ref according to Git.
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
