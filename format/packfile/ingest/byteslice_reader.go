package ingest

import "io"

// byteSliceReader implements io.ByteReader on []byte.
type byteSliceReader struct {
	data []byte
	pos  int
}

// ReadByte reads one byte from receiver.
func (reader *byteSliceReader) ReadByte() (byte, error) {
	if reader.pos >= len(reader.data) {
		return 0, io.EOF
	}

	b := reader.data[reader.pos]
	reader.pos++

	return b, nil
}
