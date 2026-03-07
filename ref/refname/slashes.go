package refname

import "strings"

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
