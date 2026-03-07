// Package utils provides misc utilities.
package utils

import (
	"fmt"
	"io"
)

// WriteProgressf writes one formatted progress message to w.
//
// It is nil-safe and ignores write errors by design.
func WriteProgressf(w io.Writer, format string, args ...any) {
	if w == nil {
		return
	}

	_, _ = fmt.Fprintf(w, format, args...)
}
