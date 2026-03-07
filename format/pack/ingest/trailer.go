package ingest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

// finishAndFlushTrailer reads trailer hash bytes, verifies trailer checksum,
// and optionally requires the source stream to hit EOF afterward.
func (scanner *streamScanner) finishAndFlushTrailer(requireTrailingEOF bool) error {
	if scanner.hashSize <= 0 {
		return fmt.Errorf("format/pack/ingest: invalid hash size")
	}

	trailer := make([]byte, scanner.hashSize)

	scanner.hashEnabled = false

	err := scanner.readFull(trailer)
	if err != nil {
		return &PackTrailerMismatchError{}
	}

	scanner.packTrailer = append(scanner.packTrailer[:0], trailer...)

	if scanner.n-scanner.off > 0 {
		return fmt.Errorf("format/pack/ingest: pack has trailing garbage")
	}

	if !requireTrailingEOF {
		computed := scanner.hash.Sum(nil)
		if !bytes.Equal(computed, trailer) {
			return &PackTrailerMismatchError{}
		}

		return nil
	}

	var probe [1]byte

	n, err := scanner.Read(probe[:])
	if n > 0 || err == nil {
		return fmt.Errorf("format/pack/ingest: pack has trailing garbage")
	}

	if !errors.Is(err, io.EOF) {
		return err
	}

	computed := scanner.hash.Sum(nil)
	if !bytes.Equal(computed, trailer) {
		return &PackTrailerMismatchError{}
	}

	return nil
}
