package pktline

import (
	"errors"
	"fmt"
	"io"
)

// ReadOptions controls decoding behavior.
type ReadOptions struct {
	// ChompLF removes one trailing '\n' from PacketData payloads.
	ChompLF bool
}

// Decoder reads pkt-line frames from an io.Reader.
//
// It preserves frame boundaries and supports one-frame lookahead via PeekFrame.
type Decoder struct {
	r       io.Reader
	maxData int
	opts    ReadOptions

	peeked  bool
	peek    Frame
	peekErr error
}

// NewDecoder creates a decoder over r.
func NewDecoder(r io.Reader, opts ReadOptions) *Decoder {
	return &Decoder{
		r:       r,
		maxData: LargePacketDataMax,
		opts:    opts,
	}
}

// SetMaxData sets maximum payload size accepted for one data packet.
//
// Non-positive n resets to LargePacketDataMax.
func (d *Decoder) SetMaxData(n int) {
	if n <= 0 {
		d.maxData = LargePacketDataMax

		return
	}

	d.maxData = n
}

func cloneFrame(f Frame) Frame {
	if f.Type != PacketData {
		return Frame{Type: f.Type}
	}

	out := Frame{Type: f.Type}
	if f.Payload != nil {
		out.Payload = append([]byte(nil), f.Payload...)
	}

	return out
}

// ReadFrame reads one frame.
//
// 0000 is a PacketFlush
// 0001 is a PacketDelim
// 0002 is a PacketResponseEnd
// 0004 is a PacketData with empty payload
//
// 0003 and malformed headers return *ProtocolError.
func (d *Decoder) ReadFrame() (Frame, error) {
	if d.peeked {
		d.peeked = false

		return cloneFrame(d.peek), d.peekErr
	}

	return d.readFrame()
}

// PeekFrame returns the next frame without consuming it.
//
// A subsequent ReadFrame returns the same frame.
func (d *Decoder) PeekFrame() (Frame, error) {
	if !d.peeked {
		d.peek, d.peekErr = d.readFrame()
		d.peeked = true
	}

	return cloneFrame(d.peek), d.peekErr
}

func (d *Decoder) readFrame() (Frame, error) {
	var hdr [4]byte

	_, err := io.ReadFull(d.r, hdr[:])
	if err != nil {
		if errors.Is(err, io.EOF) {
			return Frame{}, io.EOF
		}

		if errors.Is(err, io.ErrUnexpectedEOF) {
			return Frame{}, io.ErrUnexpectedEOF
		}

		return Frame{}, err
	}

	n, err := ParseLengthHeader(hdr)
	if err != nil {
		return Frame{}, &ProtocolError{Header: hdr, Reason: err.Error()}
	}

	switch n {
	case 0:
		return Frame{Type: PacketFlush}, nil
	case 1:
		return Frame{Type: PacketDelim}, nil
	case 2:
		return Frame{Type: PacketResponseEnd}, nil
	case 3:
		return Frame{}, &ProtocolError{Header: hdr, Reason: "invalid pkt-line length 3"}
	}

	if n < 4 {
		return Frame{}, &ProtocolError{Header: hdr, Reason: fmt.Sprintf("invalid pkt-line length %d", n)}
	}

	if n > LargePacketMax {
		perr := &ProtocolError{Header: hdr, Reason: fmt.Sprintf("pkt-line length %d exceeds max %d", n, LargePacketMax)}

		err := d.discardPayload(n - 4)
		if err != nil {
			return Frame{}, errors.Join(perr, err)
		}

		return Frame{}, perr
	}

	payloadLen := n - 4
	if payloadLen > d.maxData {
		serr := fmt.Errorf("%w: %d > %d", ErrTooLarge, payloadLen, d.maxData)

		err := d.discardPayload(payloadLen)
		if err != nil {
			return Frame{}, errors.Join(serr, err)
		}

		return Frame{}, serr
	}

	payload := make([]byte, payloadLen)

	_, err = io.ReadFull(d.r, payload)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return Frame{}, io.ErrUnexpectedEOF
		}

		return Frame{}, err
	}

	if d.opts.ChompLF && len(payload) > 0 && payload[len(payload)-1] == '\n' {
		payload = payload[:len(payload)-1]
	}

	return Frame{Type: PacketData, Payload: payload}, nil
}

func (d *Decoder) discardPayload(n int) error {
	if n <= 0 {
		return nil
	}

	_, err := io.CopyN(io.Discard, d.r, int64(n))
	if err == nil {
		return nil
	}

	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return io.ErrUnexpectedEOF
	}

	return err
}
