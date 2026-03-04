package adler32_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/internal/adler32"
)

const benchmarkSize = 64 * 1024

var data = make([]byte, benchmarkSize)

func init() { //nolint:gochecknoinits
	for i := range benchmarkSize {
		data[i] = byte(i % 256)
	}
}

func BenchmarkChecksum(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		adler32.Checksum(data)
	}
}
