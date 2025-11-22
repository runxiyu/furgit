package zlibx

import (
	"bytes"
	stdzlib "compress/zlib"
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
	payload := []byte("hello, zlib world!")
	compressed := compressZlib(t, payload)

	out, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress: %v", err)
	}
	defer out.Release()

	if !bytes.Equal(out.Bytes(), payload) {
		t.Fatalf("unexpected payload %q", out.Bytes())
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
