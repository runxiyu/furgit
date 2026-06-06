package name

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

func collapseSlashes(name string) string {
	if name == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(name))

	prev := byte('/')

	for i := range len(name) {
		ch := name[i]
		if prev == '/' && ch == '/' {
			continue
		}

		builder.WriteByte(ch)
		prev = ch
	}

	return builder.String()
}
