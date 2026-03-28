package server

import (
	"io"

	"codeberg.org/lindenii/furgit/common/iowrap"
	"codeberg.org/lindenii/furgit/network/protocol/pktline"
	"codeberg.org/lindenii/furgit/network/protocol/sideband64k"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

// Options configures one server-side v0/v1 session.
type Options struct {
	// Version selects protocol v0 or v1 framing.
	Version Version
	// Algorithm is the repository object ID algorithm for this session.
	Algorithm objectid.Algorithm
}

// Session is one stateful server-side v0/v1 server protocol session.
type Session struct {
	dec         *pktline.Decoder
	enc         *pktline.Encoder
	sideband    *sideband64k.Encoder
	opts        Options
	useSideBand bool
}

// NewSession creates one v0/v1 server session over r and w.
func NewSession(r io.Reader, w iowrap.WriteFlusher, opts Options) *Session {
	return &Session{
		dec:      pktline.NewDecoder(r, pktline.ReadOptions{}),
		enc:      pktline.NewEncoder(w),
		sideband: sideband64k.NewEncoder(w),
		opts:     opts,
	}
}

// Algorithm returns the session object ID algorithm.
func (session *Session) Algorithm() objectid.Algorithm {
	return session.opts.Algorithm
}

// ReadFrame reads one low-level pkt-line frame from the session input.
func (session *Session) ReadFrame() (Frame, error) {
	return session.dec.ReadFrame()
}

// EnableSideBand64K enables side-band-64k output framing for subsequent data,
// progress, error, and flush writes.
func (session *Session) EnableSideBand64K() {
	session.useSideBand = true
}

// WriteData writes one primary output packet.
func (session *Session) WriteData(p []byte) error {
	if session.useSideBand {
		return session.sideband.WriteData(p)
	}

	return session.enc.WriteData(p)
}

// WriteProgress writes one progress packet.
func (session *Session) WriteProgress(p []byte) error {
	if !session.useSideBand {
		return ErrSideBandNotEnabled
	}

	return session.sideband.WriteProgress(p)
}

// WriteError writes one fatal error packet.
func (session *Session) WriteError(p []byte) error {
	if !session.useSideBand {
		return ErrSideBandNotEnabled
	}

	return session.sideband.WriteError(p)
}

// WriteFlushPacket writes one trailing flush packet.
func (session *Session) WriteFlushPacket() error {
	if session.useSideBand {
		return session.sideband.WriteFlushPacket()
	}

	return session.enc.WriteFlushPacket()
}

// Flush flushes buffered transport output without emitting pkt-line frames.
func (session *Session) Flush() error {
	if session.useSideBand {
		return session.sideband.Flush()
	}

	return session.enc.Flush()
}

type flushWriter struct {
	writer io.Writer
	flush  func() error
}

func (w flushWriter) Write(p []byte) (int, error) {
	return w.writer.Write(p)
}

func (w flushWriter) Flush() error {
	return w.flush()
}

// ProgressWriter returns one chunking writer for sideband progress output.
//
// When side-band-64k was not negotiated, writes are discarded.
func (session *Session) ProgressWriter() iowrap.WriteFlusher {
	if !session.useSideBand {
		return iowrap.NopFlush(io.Discard)
	}

	return flushWriter{
		writer: sideband64k.NewChunkWriter(session.sideband, sideband64k.BandProgress),
		flush:  session.sideband.Flush,
	}
}

// ErrorWriter returns one chunking writer for sideband error output.
//
// When side-band-64k was not negotiated, writes are discarded.
func (session *Session) ErrorWriter() io.Writer {
	if !session.useSideBand {
		return io.Discard
	}

	return sideband64k.NewChunkWriter(session.sideband, sideband64k.BandError)
}

// PrimaryDataWriter returns one chunking writer for primary output bytes.
//
// When side-band-64k is enabled, writes are chunked into band-1 sideband
// frames. Otherwise writes are chunked into direct pkt-line data frames.
func (session *Session) PrimaryDataWriter() io.Writer {
	if session.useSideBand {
		return sideband64k.NewChunkWriter(session.sideband, sideband64k.BandData)
	}

	return pktline.NewChunkWriter(session.enc)
}
