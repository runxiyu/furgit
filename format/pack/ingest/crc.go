package ingest

import "fmt"

// beginEntryCRC starts inline CRC accumulation for one packed entry.
func (scanner *streamScanner) beginEntryCRC() {
	scanner.entryCRC = 0
	scanner.inEntryCRC = true
}

// endEntryCRC finishes inline CRC accumulation for one packed entry.
func (scanner *streamScanner) endEntryCRC() (uint32, error) {
	if !scanner.inEntryCRC {
		return 0, fmt.Errorf("format/pack/ingest: entry CRC not started")
	}

	crc := scanner.entryCRC
	scanner.entryCRC = 0
	scanner.inEntryCRC = false

	return crc, nil
}
