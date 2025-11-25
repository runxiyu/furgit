package zlibx

import (
	"bytes"
	stdzlib "compress/zlib"
	"crypto/rand"
	"testing"
)

func compressZlib(t *testing.T, payload []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := stdzlib.NewWriter(&buf)
	if _, err := w.Write(payload); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return buf.Bytes()
}

func TestDecompress(t *testing.T) {
	makeRand := func(n int) []byte {
		b := make([]byte, n)
		if _, err := rand.Read(b); err != nil {
			t.Fatalf("rand.Read: %v", err)
		}
		return b
	}

	type tc struct {
		name    string
		payload []byte
	}

	tests := []tc{
		{
			name:    "simple-hello",
			payload: []byte("hello, zlib world!"),
		},
		{
			name:    "empty",
			payload: []byte{},
		},
		{
			name:    "single-byte",
			payload: []byte{0x42},
		},
		{
			name:    "all-zero-1k",
			payload: bytes.Repeat([]byte{0}, 1024),
		},
		{
			name:    "all-FF-1k",
			payload: bytes.Repeat([]byte{0xFF}, 1024),
		},
		{
			name:    "ascii-repeated-pattern",
			payload: bytes.Repeat([]byte("ABC123!"), 500),
		},
		{
			name: "binary-structured",
			payload: []byte{
				0x00, 0x01, 0x02, 0x03,
				0x10, 0x20, 0x30, 0x40,
				0xFF, 0xEE, 0xDD, 0xCC,
			},
		},
		{
			name:    "1k-crypto-random",
			payload: makeRand(1024),
		},
		{
			name:    "32k-crypto-random",
			payload: makeRand(32 * 1024),
		},
		{
			name:    "256k-crypto-random",
			payload: makeRand(256 * 1024),
		},
		{
			name:    "highly-compressible-large",
			payload: bytes.Repeat([]byte("AAAAAAAAAAAAAAAAAAAA"), 50_000),
		},
		{
			name:    "json",
			payload: []byte(`{"name":"test","values":[1,2,3,4],"deep":{"x":123,"y":"abc"}}`),
		},
		{
			name:    "html",
			payload: []byte("<html><body><h1>Title</h1><p>Paragraph</p></body></html>"),
		},
		{
			name: "alternating-binary-pattern",
			payload: func() []byte {
				b := make([]byte, 4096)
				for i := 0; i < len(b); i++ {
					if i%2 == 0 {
						b[i] = 0xAA
					} else {
						b[i] = 0x55
					}
				}
				return b
			}(),
		},
		{
			name:    "large-repetitive-words",
			payload: bytes.Repeat([]byte("the quick brown fox jumps over the lazy dog\n"), 4000),
		},
		{
			name:    "unicode",
			payload: []byte("我不知道该说点啥就随便打点字吧🤷‍♀️"),
		},
		{
			name:    "multi-meg-random-2MB",
			payload: makeRand(2 * 1024 * 1024),
		},
		{
			name:    "multi-meg-random-16MB",
			payload: makeRand(16 * 1024 * 1024),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed := compressZlib(t, tt.payload)

			out, err := Decompress(compressed)
			if err != nil {
				t.Fatalf("Decompress: %v", err)
			}
			defer out.Release()

			if !bytes.Equal(out.Bytes(), tt.payload) {
				t.Fatalf("payload mismatch: got %d bytes, want %d", len(out.Bytes()), len(tt.payload))
			}
		})
	}
}

func TestDecompressChecksumError(t *testing.T) {
	payload := []byte("checksum check")
	compressed := compressZlib(t, payload)
	compressed[len(compressed)-1] ^= 0xff

	if _, err := Decompress(compressed); err != ErrChecksum {
		t.Fatalf("expected ErrChecksum, got %v", err)
	}
}

func TestDecompressSizedUsesHint(t *testing.T) {
	payload := []byte("tiny payload")
	compressed := compressZlib(t, payload)

	const hint = 1 << 20
	out, err := DecompressSized(compressed, hint)
	if err != nil {
		t.Fatalf("DecompressSized: %v", err)
	}
	defer out.Release()

	if !bytes.Equal(out.Bytes(), payload) {
		t.Fatalf("unexpected payload %q", out.Bytes())
	}
	if cap(out.Bytes()) < hint {
		t.Fatalf("expected capacity >= %d, got %d", hint, cap(out.Bytes()))
	}
}
