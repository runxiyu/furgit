package ingest

import "os"

// fileSectionWriter writes sequentially to file via WriteAt at one base offset.
type fileSectionWriter struct {
	file *os.File
	off  int64
	pos  int64
}

// Write writes src at current section position.
func (writer *fileSectionWriter) Write(src []byte) (int, error) {
	if len(src) == 0 {
		return 0, nil
	}

	n, err := writer.file.WriteAt(src, writer.off+writer.pos)
	writer.pos += int64(n)

	return n, err
}
