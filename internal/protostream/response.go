// Package protostream provides helpers for framing protocol-v2 responses.
package protostream

import (
	"io"

	"codeberg.org/lindenii/furgit/pktline"
)

// StreamCode identifies the multiplexed stream in a side-band pkt-line payload.
type StreamCode byte

const (
	StreamData     StreamCode = 1
	StreamProgress StreamCode = 2
	StreamError    StreamCode = 3
)

const (
	// maxPktLineData is the maximum pkt-line payload size (len excludes header).
	// See gitprotocol-common(5); the maximum pkt-line length is 65520, including header.
	maxPktLineData = 65516
	// maxSidebandData is max pkt-line payload minus 1 byte for the stream code.
	maxSidebandData = maxPktLineData - 1
)

// ResponseOptions configures how responses are framed.
type ResponseOptions struct {
	// NoProgress suppresses progress messages (stream code 2).
	NoProgress bool
	// SidebandAll enables side-banding for all non-flush/delim/response-end pkt-lines.
	SidebandAll bool
}

// ResponseWriter writes protocol-v2 responses with optional side-band multiplexing.
type ResponseWriter struct {
	pw   *pktline.Writer
	opts ResponseOptions
}

// NewResponseWriter returns a ResponseWriter wrapping pw.
func NewResponseWriter(pw *pktline.Writer, opts ResponseOptions) *ResponseWriter {
	return &ResponseWriter{
		pw:   pw,
		opts: opts,
	}
}

// WriteLine writes a pkt-line payload. If SidebandAll is enabled, the payload
// is sent on StreamData.
func (rw *ResponseWriter) WriteLine(payload []byte) error {
	if rw.opts.SidebandAll {
		return rw.writeStream(StreamData, payload, false)
	}
	return rw.pw.WriteLine(payload)
}

// Flush writes a flush-pkt.
func (rw *ResponseWriter) Flush() error {
	return rw.pw.Flush()
}

// Delim writes a delim-pkt.
func (rw *ResponseWriter) Delim() error {
	return rw.pw.Delim()
}

// ResponseEnd writes a response-end pkt.
func (rw *ResponseWriter) ResponseEnd() error {
	return rw.pw.ResponseEnd()
}

// PackWriter returns an io.Writer that emits pack data on StreamData.
func (rw *ResponseWriter) PackWriter() io.Writer {
	return packWriter{rw: rw}
}

// WritePack writes pack data on StreamData.
func (rw *ResponseWriter) WritePack(p []byte) error {
	if len(p) == 0 {
		return nil
	}
	return rw.writeStream(StreamData, p, false)
}

// WriteProgress writes a progress message on StreamProgress.
func (rw *ResponseWriter) WriteProgress(p []byte) error {
	if rw.opts.NoProgress || len(p) == 0 {
		return nil
	}
	return rw.writeStream(StreamProgress, p, false)
}

// WriteError writes a fatal error message on StreamError.
func (rw *ResponseWriter) WriteError(p []byte) error {
	if len(p) == 0 {
		return nil
	}
	return rw.writeStream(StreamError, p, false)
}

// Keepalive writes a side-band progress keepalive ("0005\\2").
func (rw *ResponseWriter) Keepalive() error {
	if rw.opts.NoProgress {
		return nil
	}
	return rw.writeStream(StreamProgress, nil, true)
}

func (rw *ResponseWriter) writeStream(code StreamCode, p []byte, allowEmpty bool) error {
	if !allowEmpty && len(p) == 0 {
		return nil
	}

	buf := make([]byte, 1+maxSidebandData)
	for len(p) > 0 || allowEmpty {
		n := len(p)
		if n > maxSidebandData {
			n = maxSidebandData
		}
		buf[0] = byte(code)
		if n > 0 {
			copy(buf[1:], p[:n])
		}
		if err := rw.pw.WriteLine(buf[:1+n]); err != nil {
			return err
		}
		p = p[n:]
		allowEmpty = false
	}
	return nil
}

type packWriter struct {
	rw *ResponseWriter
}

func (pw packWriter) Write(p []byte) (int, error) {
	if err := pw.rw.WritePack(p); err != nil {
		return 0, err
	}
	return len(p), nil
}
