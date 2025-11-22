package zlibx

import (
	"bytes"
	stdzlib "compress/zlib"
	"testing"
)

func compressZlib(t *testing.T, payload, dict []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	var (
		w   *stdzlib.Writer
		err error
	)
	if dict != nil {
		w, err = stdzlib.NewWriterLevelDict(&buf, stdzlib.DefaultCompression, dict)
	} else {
		w = stdzlib.NewWriter(&buf)
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
	payload := []byte("hello, zlib world!")
	compressed := compressZlib(t, payload, nil)

	out, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress: %v", err)
	}
	defer out.Release()

	if !bytes.Equal(out.Bytes(), payload) {
		t.Fatalf("unexpected payload %q", out.Bytes())
	}
}

func TestDecompressDict(t *testing.T) {
	dict := []byte("git dictionary for zlib")
	payload := append([]byte(nil), dict...)
	payload = append(payload, []byte(" -- extended body -- extended body")...)
	compressed := compressZlib(t, payload, dict)

	out, err := DecompressDict(compressed, dict)
	if err != nil {
		t.Fatalf("DecompressDict: %v", err)
	}
	defer out.Release()

	if !bytes.Equal(out.Bytes(), payload) {
		t.Fatalf("unexpected payload %q", out.Bytes())
	}
}

func TestDecompressDictMissing(t *testing.T) {
	dict := []byte("preset dictionary")
	payload := append([]byte(nil), dict...)
	payload = append(payload, []byte(" .. more data ..")...)
	compressed := compressZlib(t, payload, dict)

	if _, err := Decompress(compressed); err != ErrDictionary {
		t.Fatalf("expected ErrDictionary, got %v", err)
	}
}

func TestDecompressChecksumError(t *testing.T) {
	payload := []byte("checksum check")
	compressed := compressZlib(t, payload, nil)
	compressed[len(compressed)-1] ^= 0xff

	if _, err := Decompress(compressed); err != ErrChecksum {
		t.Fatalf("expected ErrChecksum, got %v", err)
	}
}

func TestDecompressSizedUsesHint(t *testing.T) {
	payload := []byte("tiny payload")
	compressed := compressZlib(t, payload, nil)

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
