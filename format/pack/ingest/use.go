package ingest

import (
	"fmt"
	"hash/crc32"
)

// use consumes n unread bytes and updates accounting/checksum state.
func (scanner *streamScanner) use(n int) error {
	if n < 0 || n > scanner.n-scanner.off {
		return fmt.Errorf("format/pack/ingest: invalid consume length %d", n)
	}

	if n == 0 {
		return nil
	}

	chunk := scanner.buf[scanner.off : scanner.off+n]
	if scanner.hashEnabled {
		_, err := scanner.hash.Write(chunk)
		if err != nil {
			return err
		}
	}

	if scanner.inEntryCRC {
		scanner.entryCRC = crc32.Update(scanner.entryCRC, crc32.IEEETable, chunk)
	}

	scanner.off += n
	scanner.consumed += uint64(n)

	return nil
}
