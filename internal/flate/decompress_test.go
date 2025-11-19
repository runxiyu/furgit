package flate

import (
	"bytes"
	stdflate "compress/flate"
	"testing"
)

func compressDeflate(t *testing.T, payload, dict []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	var (
		w   *stdflate.Writer
		err error
	)
	if dict != nil {
		w, err = stdflate.NewWriterDict(&buf, stdflate.DefaultCompression, dict)
	} else {
		w, err = stdflate.NewWriter(&buf, stdflate.DefaultCompression)
	}
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

func TestDecompress(t *testing.T) {
	payload := bytes.Repeat([]byte("golang"), 32)
	compressed := compressDeflate(t, payload, nil)

	out, _, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress: %v", err)
	}
	defer out.Release()

	if !bytes.Equal(out.Bytes(), payload) {
		t.Fatalf("unexpected payload: got %q", out.Bytes())
	}
}

func TestDecompressDict(t *testing.T) {
	dict := []byte("furgit dictionary payload")
	payload := append([]byte(nil), dict...)
	payload = append(payload, []byte(" -- and some more data repeated -- and some more data repeated")...)

	compressed := compressDeflate(t, payload, dict)

	out, _, err := DecompressDict(compressed, dict)
	if err != nil {
		t.Fatalf("DecompressDict: %v", err)
	}
	defer out.Release()

	if !bytes.Equal(out.Bytes(), payload) {
		t.Fatalf("unexpected payload: got %q", out.Bytes())
	}
}

func TestDecompressDictMissing(t *testing.T) {
	dict := []byte("shared prefix to enforce dictionary usage")
	payload := append([]byte(nil), dict...)
	payload = append(payload, []byte(" trailing data to force reference")...)

	compressed := compressDeflate(t, payload, dict)

	if _, _, err := Decompress(compressed); err == nil {
		t.Fatalf("expected error when dictionary missing")
	}
}
