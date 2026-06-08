package utils

import (
	"fmt"
	"io"
)

// BestEffortFprintf writes one formatted message to w.
//
// It is nil-safe and ignores write errors by design.
func BestEffortFprintf(w io.Writer, format string, args ...any) {
	if w == nil {
		return
	}

	_, _ = fmt.Fprintf(w, format, args...)
}
