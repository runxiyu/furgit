package name

import "strings"

// IsSafe reports whether name is a safe name
// for direct filesystem operations.
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

func normalizeRefPath(path string) (string, bool) {
	components := make([]string, 0, strings.Count(path, "/")+1)
	i := 0

	for i < len(path) {
		for i < len(path) && path[i] == '/' {
			i++
		}

		if i == len(path) {
			break
		}

		j := i
		for j < len(path) && path[j] != '/' {
			j++
		}

		component := path[i:j]
		switch component {
		case ".":
		case "..":
			if len(components) == 0 {
				return "", false
			}

			components = components[:len(components)-1]
		default:
			components = append(components, component)
		}

		i = j
	}

	return strings.Join(components, "/"), true
}
