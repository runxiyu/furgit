// Package utils provides misc utilities.
package utils

import (
	"fmt"
	"io"
)

// FprintfBestEffort writes one formatted message to w.
//
// It is nil-safe and ignores write errors by design.
func FprintfBestEffort(w io.Writer, format string, args ...any) {
	if w == nil {
		return
	}

	_, _ = fmt.Fprintf(w, format, args...)
}
