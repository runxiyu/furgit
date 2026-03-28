package sideband64k

import "io"

// ChunkWriter packetizes arbitrary stream bytes into side-band-64k data frames
// for one fixed band.
//
// It never writes control packets automatically.
type ChunkWriter struct {
	enc  *Encoder
	band Band
}

// NewChunkWriter creates a chunking adapter over enc for one band.
func NewChunkWriter(enc *Encoder, band Band) *ChunkWriter {
	return &ChunkWriter{enc: enc, band: band}
}

// Write splits p into sideband frames not larger than enc's maxData.
func (cw *ChunkWriter) Write(p []byte) (int, error) {
	total := 0
	maxData := cw.enc.effectiveMaxData()

	for len(p) > 0 {
		n := min(len(p), maxData)

		err := cw.enc.WriteBand(cw.band, p[:n])
		if err != nil {
			return total, err
		}

		total += n
		p = p[n:]
	}

	return total, nil
}

// ReadFrom reads from r and writes sideband frames to the encoder.
func (cw *ChunkWriter) ReadFrom(r io.Reader) (int64, error) {
	buf := make([]byte, cw.enc.effectiveMaxData())

	var total int64

	for {
		n, err := r.Read(buf)
		if n > 0 {
			werr := cw.enc.WriteBand(cw.band, buf[:n])
			if werr != nil {
				return total, werr
			}

			total += int64(n)
		}

		if err != nil {
			if err == io.EOF {
				return total, nil
			}

			return total, err
		}
	}
}

// Flush flushes buffered output in the underlying transport.
func (cw *ChunkWriter) Flush() error {
	return cw.enc.Flush()
}
