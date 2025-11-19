package adler32_test

import (
	"testing"

	"git.sr.ht/~runxiyu/furgit/internal/adler32"
)

const benchmarkSize = 64 * 1024

var data = make([]byte, benchmarkSize)

func init() {
	for i := 0; i < benchmarkSize; i++ {
		data[i] = byte(i % 256)
	}
}

func BenchmarkChecksum(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		adler32.Checksum(data)
	}
}
