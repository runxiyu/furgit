package sideband64k

import (
	"fmt"

	"codeberg.org/lindenii/furgit/format/pktline"
)

// Encoder writes side-band-64k frames to a flush-capable output transport.
//
// It writes exactly one frame per method call and does not auto-chunk data.
type Encoder struct {
	enc     *pktline.Encoder
	maxData int
}

// NewEncoder creates an encoder over w.
func NewEncoder(w pktline.WriteFlusher) *Encoder {
	return &Encoder{
		enc:     pktline.NewEncoder(w),
		maxData: DataMax,
	}
}

// SetMaxData sets the maximum payload size accepted by WriteBand.
//
// Non-positive n resets to DataMax.
func (e *Encoder) SetMaxData(n int) {
	if n <= 0 {
		e.maxData = DataMax

		return
	}

	e.maxData = n
}

// WriteBand writes one side-band-64k data frame for the given band.
func (e *Encoder) WriteBand(band Band, p []byte) error {
	if !validBand(band) {
		return fmt.Errorf("%w: %d", ErrInvalidBand, band)
	}

	maxData := e.effectiveMaxData()
	if len(p) > maxData {
		return fmt.Errorf("%w: %d > %d", ErrTooLarge, len(p), maxData)
	}

	framed := make([]byte, len(p)+1)
	framed[0] = byte(band)
	copy(framed[1:], p)

	return e.enc.WriteData(framed)
}

// WriteData writes one band-1 data frame.
func (e *Encoder) WriteData(p []byte) error {
	return e.WriteBand(BandData, p)
}

// WriteProgress writes one band-2 progress frame.
func (e *Encoder) WriteProgress(p []byte) error {
	return e.WriteBand(BandProgress, p)
}

// WriteError writes one band-3 error frame.
func (e *Encoder) WriteError(p []byte) error {
	return e.WriteBand(BandError, p)
}

// WriteFlush writes control frame 0000 (flush-pkt).
func (e *Encoder) WriteFlush() error {
	return e.enc.WriteFlush()
}

// WriteDelim writes control frame 0001 (delim-pkt).
func (e *Encoder) WriteDelim() error {
	return e.enc.WriteDelim()
}

// WriteResponseEnd writes control frame 0002 (response-end-pkt).
func (e *Encoder) WriteResponseEnd() error {
	return e.enc.WriteResponseEnd()
}

// FlushIO flushes buffered output in the underlying transport.
func (e *Encoder) FlushIO() error {
	return e.enc.FlushIO()
}

// WriteFlushAndFlushIO writes a flush-pkt (0000) then flushes transport I/O.
func (e *Encoder) WriteFlushAndFlushIO() error {
	return e.enc.WriteFlushAndFlushIO()
}

func (e *Encoder) effectiveMaxData() int {
	return effectiveMaxData(e.maxData)
}
