package flatex

import (
	"bytes"
	stdflate "compress/flate"
	"testing"
)

func compressDeflate(t *testing.T, payload []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w, err := stdflate.NewWriter(&buf, stdflate.DefaultCompression)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	if _, err := w.Write(payload); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return buf.Bytes()
}

func TestDecompressSized(t *testing.T) {
	payload := bytes.Repeat([]byte("golang"), 32)
	compressed := compressDeflate(t, payload)

	out, _, err := DecompressSized(compressed, 0)
	if err != nil {
		t.Fatalf("DecompressSized: %v", err)
	}
	defer out.Release()

	if !bytes.Equal(out.Bytes(), payload) {
		t.Fatalf("unexpected payload: got %q", out.Bytes())
	}
}

func TestDecompressSizedUsesHint(t *testing.T) {
	payload := []byte("short")
	compressed := compressDeflate(t, payload)

	const hint = 1 << 19
	out, _, err := DecompressSized(compressed, hint)
	if err != nil {
		t.Fatalf("DecompressSized: %v", err)
	}
	defer out.Release()

	if !bytes.Equal(out.Bytes(), payload) {
		t.Fatalf("unexpected payload: got %q", out.Bytes())
	}
	if cap(out.Bytes()) < hint {
		t.Fatalf("expected capacity >= %d, got %d", hint, cap(out.Bytes()))
	}
}
