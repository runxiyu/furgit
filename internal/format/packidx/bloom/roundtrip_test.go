package bloom_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"lindenii.org/go/furgit/internal/format/packidx/bloom"
	"lindenii.org/go/furgit/object/id"
)

func makeOID(size int, seed uint64) []byte {
	out := make([]byte, size)
	state := seed

	for i := 0; i < size; i += 8 {
		state = state*6364136223846793005 + 1442695040888963407

		var word [8]byte

		binary.BigEndian.PutUint64(word[:], state)
		copy(out[i:], word[:])
	}

	return out
}

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	for _, format := range id.SupportedObjectFormats() {
		t.Run(format.String(), func(t *testing.T) {
			t.Parallel()

			const objects = 10000

			bucketCount, k, err := bloom.RecommendParams(format, objects)
			if err != nil {
				t.Fatal(err)
			}

			size := format.Size()
			packHash := makeOID(size, 0xC0FFEE)

			builder, err := bloom.NewBuilder(format, bucketCount, k, packHash)
			if err != nil {
				t.Fatal(err)
			}

			for i := range objects {
				builder.Add(makeOID(size, uint64(i))) //#nosec G115
			}

			filter, err := bloom.Parse(builder.Bytes(), format)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(filter.PackHash(), packHash) {
				t.Fatalf("PackHash = %x, want %x", filter.PackHash(), packHash)
			}

			err = filter.Verify()
			if err != nil {
				t.Fatalf("Verify on a freshly built filter: %v", err)
			}

			for i := range objects {
				if !filter.MayContain(makeOID(size, uint64(i))) { //#nosec G115
					t.Fatalf("false negative for added object %d", i)
				}
			}

			const probes = 10000

			falsePositives := 0

			for i := range probes {
				if filter.MayContain(makeOID(size, uint64(1)<<40+uint64(i))) { //#nosec G115
					falsePositives++
				}
			}

			rate := float64(falsePositives) / float64(probes)
			if rate > 0.05 {
				t.Errorf("false positive rate %.4f exceeds 0.05", rate)
			}

			t.Logf("B=%d K=%d false positive rate %.4f", bucketCount, k, rate)
		})
	}
}
