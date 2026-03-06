package ingest

import "fmt"

// flush writes all consumed-but-unflushed bytes to destination pack file.
func (scanner *streamScanner) flush() error {
	return scanner.flushConsumedPrefix()
}

// flushConsumedPrefix writes scanner.buf[:scanner.off] and compacts unread
// bytes to the start of buffer.
func (scanner *streamScanner) flushConsumedPrefix() error {
	if scanner.off == 0 {
		return nil
	}

	written := 0
	for written < scanner.off {
		n, err := scanner.dstFile.Write(scanner.buf[written:scanner.off])
		if err != nil {
			return &ErrDestinationWrite{Op: fmt.Sprintf("write pack: %v", err)}
		}

		if n == 0 {
			return &ErrDestinationWrite{Op: "write pack: short write"}
		}

		written += n
	}

	unread := scanner.n - scanner.off
	copy(scanner.buf[:unread], scanner.buf[scanner.off:scanner.n])
	scanner.off = 0
	scanner.n = unread

	return nil
}
