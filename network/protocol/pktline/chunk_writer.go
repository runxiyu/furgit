package pktline

import "io"

// ChunkWriter packetizes arbitrary stream bytes into data pkt-lines.
// It never writes control packets automatically.
type ChunkWriter struct {
	enc *Encoder
}

// NewChunkWriter creates a chunking adapter over enc.
//
// Labels: Deps-Borrowed, Life-Parent.
func NewChunkWriter(enc *Encoder) *ChunkWriter {
	return &ChunkWriter{enc: enc}
}

// Write splits p into data frames not larger than enc's maxData.
//
// It implements io.Writer.
func (cw *ChunkWriter) Write(p []byte) (int, error) {
	total := 0
	maxData := cw.enc.effectiveMaxData()

	for len(p) > 0 {
		n := min(len(p), maxData)

		err := cw.enc.WriteData(p[:n])
		if err != nil {
			return total, err
		}

		total += n
		p = p[n:]
	}

	return total, nil
}

// ReadFrom reads from r and writes pkt-line data frames to the encoder.
//
// It implements io.ReaderFrom.
func (cw *ChunkWriter) ReadFrom(r io.Reader) (int64, error) {
	buf := make([]byte, cw.enc.effectiveMaxData())

	var total int64

	for {
		n, err := r.Read(buf)
		if n > 0 {
			werr := cw.enc.WriteData(buf[:n])
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
