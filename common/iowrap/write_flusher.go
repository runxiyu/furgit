package iowrap

import "io"

// WriteFlusher writes bytes and flushes buffered output state.
type WriteFlusher interface {
	io.Writer
	Flush() error
}
