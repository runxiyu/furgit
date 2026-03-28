package iowrap

import "io"

type nopFlusher struct {
	io.Writer
}

// NopFlush adapts writer into a WriteFlusher with a no-op Flush.
func NopFlush(writer io.Writer) WriteFlusher {
	if writer == nil {
		return nil
	}

	return nopFlusher{Writer: writer}
}

func (nopFlusher) Flush() error {
	return nil
}
