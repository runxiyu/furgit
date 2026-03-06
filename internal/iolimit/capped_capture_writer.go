package iolimit

import "bytes"

// CappedCaptureWriter captures written bytes up to a fixed limit.
//
// Once the total written bytes would exceed the limit, capture is disabled and
// Bytes() returns nil. Write still reports success for the full input length.
type CappedCaptureWriter struct {
	limit int64
	buf   bytes.Buffer
	full  bool
}

// NewCappedCaptureWriter constructs one capped capture writer.
func NewCappedCaptureWriter(limit int64) *CappedCaptureWriter {
	return &CappedCaptureWriter{limit: limit}
}

// Write captures up to the configured limit and always reports len(src) bytes written.
func (writer *CappedCaptureWriter) Write(src []byte) (int, error) {
	if writer.full {
		return len(src), nil
	}

	room := writer.limit - int64(writer.buf.Len())
	if room <= 0 {
		writer.full = true

		return len(src), nil
	}

	if int64(len(src)) > room {
		_, _ = writer.buf.Write(src[:room])
		writer.full = true

		return len(src), nil
	}

	_, _ = writer.buf.Write(src)

	return len(src), nil
}

// Bytes returns captured bytes, or nil when capture exceeded the limit.
func (writer *CappedCaptureWriter) Bytes() []byte {
	if writer.full {
		return nil
	}

	return writer.buf.Bytes()
}
