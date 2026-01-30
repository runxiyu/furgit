// Package pktline provides support for the pkt-line format described in gitprotocol-common(5).
package pktline

import (
	"errors"
	"io"
)

const (
	maxPacketSize    = 65520
	maxPacketDataLen = maxPacketSize - 4
)

var (
	ErrInvalidHeader  = errors.New("pktline: invalid header")
	ErrPacketTooLarge = errors.New("pktline: packet too large")
	ErrBufferTooSmall = errors.New("pktline: buffer too small")
)

type Status uint8

const (
	StatusEOF Status = iota
	StatusData
	StatusFlush
	StatusDelim
	StatusResponseEnd
)

// Reader reads pkt-line data from an io.Reader.
type Reader struct {
	r io.Reader
}

// NewReader returns a Reader that reads from r.
func NewReader(r io.Reader) *Reader {
	return &Reader{r: r}
}

// ReadLine reads a single pkt-line from the underlying reader into buf.
// It returns the payload slice, number of payload bytes, and a status.
func (pr *Reader) ReadLine(buf []byte) ([]byte, int, Status, error) {
	var header [4]byte
	if _, err := io.ReadFull(pr.r, header[:]); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, 0, StatusEOF, io.EOF
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, 0, StatusEOF, io.ErrUnexpectedEOF
		}
		return nil, 0, StatusEOF, err
	}

	n, err := parseHeader(header[:])
	if err != nil {
		return nil, 0, StatusEOF, err
	}
	switch n {
	case 0:
		return nil, 0, StatusFlush, nil
	case 1:
		return nil, 0, StatusDelim, nil
	case 2:
		return nil, 0, StatusResponseEnd, nil
	}
	if n < 4 {
		return nil, 0, StatusEOF, ErrInvalidHeader
	}
	n -= 4
	if n > maxPacketDataLen {
		return nil, 0, StatusEOF, ErrPacketTooLarge
	}
	if n > len(buf) {
		return nil, 0, StatusEOF, ErrBufferTooSmall
	}
	if _, err := io.ReadFull(pr.r, buf[:n]); err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, 0, StatusEOF, io.ErrUnexpectedEOF
		}
		return nil, 0, StatusEOF, err
	}
	return buf[:n], n, StatusData, nil
}

// Writer writes pkt-line data to an io.Writer.
type Writer struct {
	w io.Writer
}

// NewWriter returns a Writer that writes to w.
func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

// WriteLine writes a single pkt-line with data as its payload.
func (pw *Writer) WriteLine(data []byte) error {
	if len(data) > maxPacketDataLen {
		return ErrPacketTooLarge
	}
	var header [4]byte
	setHeader(header[:], len(data)+4)
	if _, err := pw.w.Write(header[:]); err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	_, err := pw.w.Write(data)
	return err
}

// Flush writes a flush-pkt ("0000").
func (pw *Writer) Flush() error {
	_, err := io.WriteString(pw.w, "0000")
	return err
}

// Delim writes a delim-pkt ("0001").
func (pw *Writer) Delim() error {
	_, err := io.WriteString(pw.w, "0001")
	return err
}

// ResponseEnd writes a response-end pkt ("0002").
func (pw *Writer) ResponseEnd() error {
	_, err := io.WriteString(pw.w, "0002")
	return err
}

func parseHeader(b []byte) (int, error) {
	if len(b) < 4 {
		return 0, ErrInvalidHeader
	}
	v0, ok := hexVal(b[0])
	if !ok {
		return 0, ErrInvalidHeader
	}
	v1, ok := hexVal(b[1])
	if !ok {
		return 0, ErrInvalidHeader
	}
	v2, ok := hexVal(b[2])
	if !ok {
		return 0, ErrInvalidHeader
	}
	v3, ok := hexVal(b[3])
	if !ok {
		return 0, ErrInvalidHeader
	}
	return (v0 << 12) | (v1 << 8) | (v2 << 4) | v3, nil
}

func setHeader(buf []byte, size int) {
	const hex = "0123456789abcdef"
	buf[0] = hex[(size>>12)&0x0f]
	buf[1] = hex[(size>>8)&0x0f]
	buf[2] = hex[(size>>4)&0x0f]
	buf[3] = hex[size&0x0f]
}

// IIRC strconv.ParseUint, encoding/hex.Decode, etc., allocate memory.
func hexVal(b byte) (int, bool) {
	switch {
	case b >= '0' && b <= '9':
		return int(b - '0'), true
	case b >= 'a' && b <= 'f':
		return int(b-'a') + 10, true
	case b >= 'A' && b <= 'F':
		return int(b-'A') + 10, true
	default:
		return 0, false
	}
}
