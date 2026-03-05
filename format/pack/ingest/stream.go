package ingest

import (
	"bufio"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"os"
)

// streamCopier reads bytes from src, writes them to packFile, and updates
// trailer verification state.
type streamCopier struct {
	reader   *bufio.Reader
	writer   *bufio.Writer
	verifier *trailerVerifier
	offset   uint64
	entryCRC hash.Hash32
}

// newStreamCopier constructs one stream copier.
func newStreamCopier(src io.Reader, packFile *os.File, verifier *trailerVerifier) *streamCopier {
	return &streamCopier{
		reader:   bufio.NewReaderSize(src, 64<<10),
		writer:   bufio.NewWriterSize(packFile, 256<<10),
		verifier: verifier,
	}
}

// Read implements io.Reader.
func (stream *streamCopier) Read(dst []byte) (int, error) {
	n, err := stream.reader.Read(dst)
	if n > 0 {
		if writeErr := stream.writeChunk(dst[:n]); writeErr != nil {
			return 0, writeErr
		}
	}

	return n, err
}

// ReadByte implements io.ByteReader.
func (stream *streamCopier) ReadByte() (byte, error) {
	b, err := stream.reader.ReadByte()
	if err != nil {
		return 0, err
	}

	if writeErr := stream.writeChunk([]byte{b}); writeErr != nil {
		return 0, writeErr
	}

	return b, nil
}

// readFull reads exactly len(dst) bytes through stream.
func (stream *streamCopier) readFull(dst []byte) error {
	_, err := io.ReadFull(stream, dst)
	if err != nil {
		return err
	}

	return nil
}

// writeChunk mirrors src bytes to destination artifacts and accounting.
func (stream *streamCopier) writeChunk(src []byte) error {
	n, err := stream.writer.Write(src)
	if err != nil {
		return &ErrDestinationWrite{Op: fmt.Sprintf("write pack: %v", err)}
	}
	if n != len(src) {
		return &ErrDestinationWrite{Op: "write pack: short write"}
	}

	if stream.entryCRC != nil {
		_, _ = stream.entryCRC.Write(src)
	}
	stream.verifier.write(src)
	stream.offset += uint64(len(src))

	return nil
}

// flush flushes buffered pack output bytes to the destination file.
func (stream *streamCopier) flush() error {
	err := stream.writer.Flush()
	if err != nil {
		return &ErrDestinationWrite{Op: fmt.Sprintf("flush pack: %v", err)}
	}

	return nil
}

// beginEntryCRC starts inline CRC accumulation for one packed entry.
func (stream *streamCopier) beginEntryCRC() {
	stream.entryCRC = crc32.NewIEEE()
}

// endEntryCRC finishes inline CRC accumulation for one packed entry.
func (stream *streamCopier) endEntryCRC() (uint32, error) {
	if stream.entryCRC == nil {
		return 0, fmt.Errorf("format/pack/ingest: entry CRC not started")
	}
	crc := stream.entryCRC.Sum32()
	stream.entryCRC = nil

	return crc, nil
}
