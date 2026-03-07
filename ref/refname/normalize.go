package refname

import "strings"

// Normalize collapses slashes according to what Git wants
// then validates the normalized name.
func Normalize(name string, options Options) (string, error) {
	normalized := collapseSlashes(name)

	err := validate(normalized, options.flags())
	if err != nil {
		return "", err
	}

	return normalized, nil
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
