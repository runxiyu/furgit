package sideband64k_test

import (
	"bytes"
	"io"
)

type limitWriter struct {
	buf         bytes.Buffer
	maxPerWrite int
	flushes     int
	shortWrite  bool
}

func (w *limitWriter) Write(p []byte) (int, error) {
	if w.shortWrite {
		return 0, nil
	}

	if w.maxPerWrite > 0 && len(p) > w.maxPerWrite {
		p = p[:w.maxPerWrite]
	}

	return w.buf.Write(p)
}

func (w *limitWriter) Flush() error {
	w.flushes++
	return nil
}

type byteReader struct {
	data []byte
}

func (r *byteReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}

	p[0] = r.data[0]
	r.data = r.data[1:]
	return 1, nil
}
