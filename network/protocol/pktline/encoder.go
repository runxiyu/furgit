package pktline

import (
	"fmt"
	"io"
)

// WriteFlusher is the output transport contract required by Encoder.
//
// Write emits framed bytes and Flush pushes buffered transport state.
type WriteFlusher interface {
	io.Writer
	Flush() error
}

// Encoder writes pkt-line frames to a flush-capable output transport.
//
// It writes exactly one frame per method call and does not auto-chunk data.
type Encoder struct {
	w       WriteFlusher
	maxData int
}

// NewEncoder creates an encoder over w.
func NewEncoder(w WriteFlusher) *Encoder {
	return &Encoder{
		w:       w,
		maxData: LargePacketDataMax,
	}
}

// SetMaxData sets the maximum payload size accepted by WriteData.
//
// Non-positive n resets to LargePacketDataMax.
func (e *Encoder) SetMaxData(n int) {
	if n <= 0 {
		e.maxData = LargePacketDataMax

		return
	}

	e.maxData = n
}

func writeAll(w io.Writer, b []byte) error {
	for len(b) > 0 {
		n, err := w.Write(b)
		if err != nil {
			return err
		}

		if n <= 0 {
			return io.ErrShortWrite
		}

		b = b[n:]
	}

	return nil
}

// WriteData writes one data frame.
//
// Empty payload is encoded as 0004.
func (e *Encoder) WriteData(p []byte) error {
	maxData := e.effectiveMaxData()
	if len(p) > maxData {
		return fmt.Errorf("%w: %d > %d", ErrTooLarge, len(p), maxData)
	}

	var hdr [4]byte

	err := EncodeLengthHeader(&hdr, len(p)+4)
	if err != nil {
		return err
	}

	err = writeAll(e.w, hdr[:])
	if err != nil {
		return err
	}

	return writeAll(e.w, p)
}

// WriteString writes one data frame containing s and returns len(s) on success.
func (e *Encoder) WriteString(s string) (int, error) {
	err := e.WriteData([]byte(s))
	if err != nil {
		return 0, err
	}

	return len(s), nil
}

// WriteFlush writes control frame 0000 (flush-pkt).
func (e *Encoder) WriteFlush() error {
	return e.writeControl(0)
}

// WriteDelim writes control frame 0001 (delim-pkt).
func (e *Encoder) WriteDelim() error {
	return e.writeControl(1)
}

// WriteResponseEnd writes control frame 0002 (response-end-pkt).
func (e *Encoder) WriteResponseEnd() error {
	return e.writeControl(2)
}

// FlushIO flushes buffered output in the underlying transport.
//
// FlushIO does not emit any pkt-line control frame.
func (e *Encoder) FlushIO() error {
	return e.w.Flush()
}

// WriteFlushAndFlushIO writes a flush-pkt (0000) then flushes transport I/O.
func (e *Encoder) WriteFlushAndFlushIO() error {
	err := e.WriteFlush()
	if err != nil {
		return err
	}

	return e.FlushIO()
}

func (e *Encoder) writeControl(n int) error {
	var hdr [4]byte

	err := EncodeLengthHeader(&hdr, n)
	if err != nil {
		return err
	}

	return writeAll(e.w, hdr[:])
}

func (e *Encoder) effectiveMaxData() int {
	if e.maxData <= 0 || e.maxData > LargePacketDataMax {
		return LargePacketDataMax
	}

	return e.maxData
}
