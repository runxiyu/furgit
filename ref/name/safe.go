package refname

import "strings"

// IsSafe reports whether name is one safe refname for direct filesystem
// operations; see refname_is_safe.
func IsSafe(name string) bool {
	rest, ok := strings.CutPrefix(name, "refs/")
	if ok {
		if rest == "" || rest[0] == '/' || rest[len(rest)-1] == '/' {
			return false
		}

		normalized, normOK := normalizeRefPath(rest)

		return normOK && normalized == rest
	}

	if name == "" {
		return false
	}

	for i := range len(name) {
		ch := name[i]
		if (ch < 'A' || ch > 'Z') && ch != '_' {
			return false
		}
	}

	return true
}
