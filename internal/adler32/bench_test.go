package adler32

import (
	"testing"
)

const benchmarkSize = 64 * 1024

var data = make([]byte, benchmarkSize)

func init() {
	for i := range benchmarkSize {
		data[i] = byte(i % 256)
	}
}

func BenchmarkChecksum(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		Checksum(data)
	}
}
