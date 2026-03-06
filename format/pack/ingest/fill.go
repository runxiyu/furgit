package ingest

import (
	"errors"
	"fmt"
	"io"
)

// fill ensures at least min unread bytes are available in receiver's buffer.
func (scanner *streamScanner) fill(minLen int) error {
	if minLen <= 0 {
		return nil
	}

	if minLen > len(scanner.buf) {
		return fmt.Errorf("format/pack/ingest: fill(%d) exceeds scanner buffer", minLen)
	}

	for scanner.n-scanner.off < minLen {
		err := scanner.flushConsumedPrefix()
		if err != nil {
			return err
		}

		readN, err := scanner.src.Read(scanner.buf[scanner.n:])
		if readN > 0 {
			scanner.n += readN
		}

		if err != nil {
			if errors.Is(err, io.EOF) && scanner.n-scanner.off >= minLen {
				return nil
			}

			return err
		}

		if readN == 0 {
			return io.ErrNoProgress
		}
	}

	return nil
}
