package ingest

import (
	"bytes"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"os"
)

const streamScannerBufferSize = 64 << 10

// streamScanner incrementally reads/consumes one pack stream while mirroring
// consumed bytes into one destination pack file.
type streamScanner struct {
	src     io.Reader
	dstFile *os.File

	// Input buffer window: buf[off:n] is unread.
	buf []byte
	off int
	n   int

	// Absolute consumed stream bytes.
	consumed uint64

	// Running pack hash over consumed bytes while hashEnabled is true.
	hash        hash.Hash
	hashSize    int
	hashEnabled bool

	// Entry CRC state while one entry is being consumed.
	entryCRC   uint32
	inEntryCRC bool

	packTrailer []byte
}

// newStreamScanner constructs one scanner with fixed input buffering.
func newStreamScanner(src io.Reader, dstFile *os.File, hash hash.Hash, hashSize int) *streamScanner {
	return &streamScanner{
		src:         src,
		dstFile:     dstFile,
		buf:         make([]byte, streamScannerBufferSize),
		hash:        hash,
		hashSize:    hashSize,
		hashEnabled: true,
	}
}

// fill ensures at least min unread bytes are available in receiver's buffer.
func (scanner *streamScanner) fill(min int) error {
	if min <= 0 {
		return nil
	}
	if min > len(scanner.buf) {
		return fmt.Errorf("format/pack/ingest: fill(%d) exceeds scanner buffer", min)
	}

	for scanner.n-scanner.off < min {
		if err := scanner.flushConsumedPrefix(); err != nil {
			return err
		}

		readN, err := scanner.src.Read(scanner.buf[scanner.n:])
		if readN > 0 {
			scanner.n += readN
		}
		if err != nil {
			if err == io.EOF && scanner.n-scanner.off >= min {
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
		if _, err := scanner.hash.Write(chunk); err != nil {
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

// Read implements io.Reader.
func (scanner *streamScanner) Read(dst []byte) (int, error) {
	if len(dst) == 0 {
		return 0, nil
	}

	if scanner.n-scanner.off == 0 {
		if err := scanner.fill(1); err != nil {
			if err == io.EOF {
				return 0, io.EOF
			}

			return 0, err
		}
	}

	unread := scanner.n - scanner.off
	if unread == 0 {
		return 0, io.EOF
	}

	n := len(dst)
	if n > unread {
		n = unread
	}
	copy(dst, scanner.buf[scanner.off:scanner.off+n])
	if err := scanner.use(n); err != nil {
		return 0, err
	}

	return n, nil
}

// ReadByte implements io.ByteReader without allocation.
func (scanner *streamScanner) ReadByte() (byte, error) {
	if scanner.n-scanner.off == 0 {
		if err := scanner.fill(1); err != nil {
			return 0, err
		}
	}

	b := scanner.buf[scanner.off]
	if err := scanner.use(1); err != nil {
		return 0, err
	}

	return b, nil
}

// readFull reads exactly len(dst) bytes through receiver.
func (scanner *streamScanner) readFull(dst []byte) error {
	_, err := io.ReadFull(scanner, dst)
	if err != nil {
		return err
	}

	return nil
}

// flush writes all consumed-but-unflushed bytes to destination pack file.
func (scanner *streamScanner) flush() error {
	return scanner.flushConsumedPrefix()
}

// finishAndFlushTrailer reads trailer hash bytes, verifies trailer checksum,
// and ensures no trailing garbage remains in stream.
func (scanner *streamScanner) finishAndFlushTrailer() error {
	if scanner.hashSize <= 0 {
		return fmt.Errorf("format/pack/ingest: invalid hash size")
	}

	trailer := make([]byte, scanner.hashSize)
	scanner.hashEnabled = false
	if err := scanner.readFull(trailer); err != nil {
		return &ErrPackTrailerMismatch{}
	}
	scanner.packTrailer = append(scanner.packTrailer[:0], trailer...)

	var probe [1]byte
	n, err := scanner.Read(probe[:])
	if n > 0 || err == nil {
		return fmt.Errorf("format/pack/ingest: pack has trailing garbage")
	}
	if err != io.EOF {
		return err
	}

	computed := scanner.hash.Sum(nil)
	if !bytes.Equal(computed, trailer) {
		return &ErrPackTrailerMismatch{}
	}

	return nil
}

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
