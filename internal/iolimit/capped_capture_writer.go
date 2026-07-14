package iolimit

import (
	"bytes"
	"fmt"

	"lindenii.org/go/lgo/intconv"
)

// CappedCaptureWriter captures written bytes up to a fixed limit.
//
// Once the total written bytes would exceed the limit,
// capture is disabled and Bytes returns nil.
// Write still reports success for the full input length.
type CappedCaptureWriter struct {
	limit uint64
	buf   bytes.Buffer //exhaustruct:optional
	full  bool         //exhaustruct:optional
}

// NewCappedCaptureWriter constructs one capped capture writer.
func NewCappedCaptureWriter(limit uint64) *CappedCaptureWriter {
	return &CappedCaptureWriter{limit: limit}
}

// Write captures up to the configured limit
// and always reports len(src) bytes written.
func (writer *CappedCaptureWriter) Write(src []byte) (int, error) {
	if writer.full {
		return len(src), nil
	}

	used, err := intconv.IntToUint64(writer.buf.Len())
	if err != nil {
		return 0, fmt.Errorf("iolimit: %w", err)
	}

	if used >= writer.limit {
		writer.full = true

		return len(src), nil
	}

	room := writer.limit - used
	if uint64(len(src)) > room {
		take, err := intconv.Uint64ToInt(room)
		if err != nil {
			return 0, fmt.Errorf("iolimit: %w", err)
		}

		_, _ = writer.buf.Write(src[:take])
		writer.full = true

		return len(src), nil
	}

	_, _ = writer.buf.Write(src)

	return len(src), nil
}

// Bytes returns captured bytes,
// or nil when capture exceeded the limit.
func (writer *CappedCaptureWriter) Bytes() []byte {
	if writer.full {
		return nil
	}

	return writer.buf.Bytes()
}
