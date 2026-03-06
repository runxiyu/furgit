package ingest

import (
	"errors"
	"hash"
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

// Read implements io.Reader.
func (scanner *streamScanner) Read(dst []byte) (int, error) {
	if len(dst) == 0 {
		return 0, nil
	}

	if scanner.n-scanner.off == 0 {
		err := scanner.fill(1)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return 0, io.EOF
			}

			return 0, err
		}
	}

	unread := scanner.n - scanner.off
	if unread == 0 {
		return 0, io.EOF
	}

	n := min(len(dst), unread)

	copy(dst, scanner.buf[scanner.off:scanner.off+n])

	err := scanner.use(n)
	if err != nil {
		return 0, err
	}

	return n, nil
}

// ReadByte implements io.ByteReader without allocation.
func (scanner *streamScanner) ReadByte() (byte, error) {
	if scanner.n-scanner.off == 0 {
		err := scanner.fill(1)
		if err != nil {
			return 0, err
		}
	}

	b := scanner.buf[scanner.off]

	err := scanner.use(1)
	if err != nil {
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
