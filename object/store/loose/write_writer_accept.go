package loose

import (
	"bytes"
	"errors"

	objectheader "lindenii.org/go/furgit/object/header"
)

// acceptFull validates and accounts raw full-object input.
func (writer *streamWriter) acceptFull(src []byte) error {
	if !writer.headerDone {
		nul := bytes.IndexByte(src, 0)
		if nul >= 0 {
			headerChunkLen := nul + 1
			writer.headerBuf = append(writer.headerBuf, src[:headerChunkLen]...)

			_, size, _, ok := objectheader.Parse(writer.headerBuf)
			if !ok {
				return errors.New("objectstore/loose: malformed object header")
			}

			writer.headerDone = true
			writer.expectedContentLeft = size

			return writer.acceptContent(int64(len(src) - headerChunkLen))
		}

		writer.headerBuf = append(writer.headerBuf, src...)

		return nil
	}

	return writer.acceptContent(int64(len(src)))
}

// acceptContent validates and accounts content byte counts.
func (writer *streamWriter) acceptContent(n int64) error {
	if n > writer.expectedContentLeft {
		return errors.New("objectstore/loose: object content exceeds declared size")
	}

	writer.expectedContentLeft -= n

	return nil
}

// writeRawChunk forwards raw bytes to the hash and deflate pipeline.
func (writer *streamWriter) writeRawChunk(src []byte) error {
	_, err := writer.hash.Write(src)
	if err != nil {
		return err
	}

	_, err = writer.zw.Write(src)
	if err != nil {
		return err
	}

	return nil
}
