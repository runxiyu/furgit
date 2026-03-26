package refname

import "strings"

func sanitizedLen(builder *strings.Builder) int {
	if builder == nil {
		return 0
	}

	return builder.Len()
}
