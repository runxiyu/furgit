package sideband64k

import (
	"fmt"
	"io"

	"codeberg.org/lindenii/furgit/network/protocol/pktline"
)

// ReadOptions controls sideband decoding behavior.
type ReadOptions struct {
	// ChompLF removes one trailing '\n' from FrameData payloads only.
	ChompLF bool
}

// Decoder reads side-band-64k frames from an io.Reader.
//
// It preserves frame boundaries and supports one-frame lookahead via
// PeekFrame.
type Decoder struct {
	dec     *pktline.Decoder
	maxData int
	opts    ReadOptions

	peeked  bool
	peek    Frame
	peekErr error
}

// NewDecoder creates a decoder over r.
//
// Labels: Deps-Borrowed.
func NewDecoder(r io.Reader, opts ReadOptions) *Decoder {
	d := &Decoder{
		dec:     pktline.NewDecoder(r, pktline.ReadOptions{}),
		maxData: DataMax,
		opts:    opts,
	}
	d.dec.SetMaxData(pktline.LargePacketDataMax)

	return d
}

// SetMaxData sets maximum payload size accepted for one sideband data packet.
//
// Non-positive n resets to DataMax.
func (d *Decoder) SetMaxData(n int) {
	if n <= 0 {
		d.maxData = DataMax

		return
	}

	d.maxData = n
}

// ReadFrame reads one frame.
func (d *Decoder) ReadFrame() (Frame, error) {
	if d.peeked {
		d.peeked = false

		return cloneFrame(d.peek), d.peekErr
	}

	return d.readFrame()
}

// PeekFrame returns the next frame without consuming it.
func (d *Decoder) PeekFrame() (Frame, error) {
	if !d.peeked {
		d.peek, d.peekErr = d.readFrame()
		d.peeked = true
	}

	return cloneFrame(d.peek), d.peekErr
}

func (d *Decoder) readFrame() (Frame, error) {
	f, err := d.dec.ReadFrame()
	if err != nil {
		return Frame{}, err
	}

	switch f.Type {
	case pktline.PacketFlush:
		return Frame{Type: FrameFlush}, nil
	case pktline.PacketDelim:
		return Frame{Type: FrameDelim}, nil
	case pktline.PacketResponseEnd:
		return Frame{Type: FrameResponseEnd}, nil
	case pktline.PacketData:
		if len(f.Payload) == 0 {
			return Frame{}, &ProtocolError{Reason: "missing sideband designator"}
		}

		payload := f.Payload[1:]
		if len(payload) > d.effectiveMaxData() {
			return Frame{}, fmt.Errorf("%w: %d > %d", ErrTooLarge, len(payload), d.effectiveMaxData())
		}

		band := Band(f.Payload[0])
		if !validBand(band) {
			return Frame{}, &ProtocolError{Reason: fmt.Sprintf("%v: %d", ErrInvalidBand, band)}
		}

		payload = append([]byte(nil), payload...)
		if d.opts.ChompLF && band == BandData && len(payload) > 0 && payload[len(payload)-1] == '\n' {
			payload = payload[:len(payload)-1]
		}

		return Frame{
			Type:    frameTypeForBand(band),
			Payload: payload,
		}, nil
	default:
		return Frame{}, &ProtocolError{Reason: "unknown pkt-line frame type"}
	}
}

func (d *Decoder) effectiveMaxData() int {
	return effectiveMaxData(d.maxData)
}

func cloneFrame(f Frame) Frame {
	if f.Type == FrameFlush || f.Type == FrameDelim || f.Type == FrameResponseEnd {
		return Frame{Type: f.Type}
	}

	out := Frame{Type: f.Type}
	if f.Payload != nil {
		out.Payload = append([]byte(nil), f.Payload...)
	}

	return out
}

func validBand(band Band) bool {
	return band == BandData || band == BandProgress || band == BandError
}

func frameTypeForBand(band Band) FrameType {
	switch band {
	case BandData:
		return FrameData
	case BandProgress:
		return FrameProgress
	case BandError:
		return FrameError
	default:
		panic("invalid sideband64k band")
	}
}

func effectiveMaxData(n int) int {
	if n <= 0 || n > DataMax {
		return DataMax
	}

	return n
}
