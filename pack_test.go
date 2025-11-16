package furgit

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"testing"
)

func compressBytes(t *testing.T, payload []byte) []byte {
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	if _, err := zw.Write(payload); err != nil {
		t.Fatalf("compress write: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("compress close: %v", err)
	}
	return buf.Bytes()
}

func TestPackSectionInflate(t *testing.T) {
	payload := []byte("pack payload")
	compressed := compressBytes(t, payload)
	body, err := packSectionInflate(bytes.NewReader(compressed), len(payload))
	if err != nil {
		t.Fatalf("packSectionInflate error: %v", err)
	}
	if got := string(body.Bytes()); got != string(payload) {
		t.Fatalf("unexpected inflated data: %q", got)
	}
	body.Release()

	body, err = packSectionInflate(bytes.NewReader(compressed), 0)
	if err != nil {
		t.Fatalf("packSectionInflate streaming error: %v", err)
	}
	if got := string(body.Bytes()); got != string(payload) {
		t.Fatalf("unexpected streaming data: %q", got)
	}
	body.Release()
}

func encodePackHeader(ty ObjectType, size int) []byte {
	first := byte((ty & 0x7) << 4)
	first |= byte(size & 0x0f)
	size >>= 4
	if size == 0 {
		return []byte{first}
	}
	first |= 0x80
	out := []byte{first}
	for size > 0 {
		b := byte(size & 0x7f)
		size >>= 7
		if size != 0 {
			b |= 0x80
		}
		out = append(out, b)
	}
	return out
}

func TestPackHeaderRead(t *testing.T) {
	buf := encodePackHeader(ObjectTypeTree, 0x1fff)
	ty, size, err := packHeaderRead(bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("packHeaderRead error: %v", err)
	}
	if ty != ObjectTypeTree || size != 0x1fff {
		t.Fatalf("unexpected header decode ty=%d size=%d", ty, size)
	}
	if _, _, err := packHeaderRead(bytes.NewReader([]byte{0x80})); err == nil {
		t.Fatal("expected error for truncated header")
	}
}

func encodeVarint(value int) []byte {
	var out []byte
	for {
		b := byte(value & 0x7f)
		value >>= 7
		if value != 0 {
			b |= 0x80
		}
		out = append(out, b)
		if value == 0 {
			break
		}
	}
	return out
}

func TestPackVarintRead(t *testing.T) {
	buf := encodeVarint(0x3456)
	pos := 0
	val, err := packVarintRead(buf, &pos)
	if err != nil {
		t.Fatalf("packVarintRead error: %v", err)
	}
	if val != 0x3456 {
		t.Fatalf("unexpected varint value: %d", val)
	}
	if pos != len(buf) {
		t.Fatalf("expected pos %d, got %d", len(buf), pos)
	}
	bad := []byte{0x80}
	pos = 0
	if _, err := packVarintRead(bad, &pos); err == nil {
		t.Fatal("expected error for unterminated varint")
	}
}

func TestPackDeltaApply(t *testing.T) {
	base := borrowedFromOwned([]byte("abcdefghij"))
	defer base.Release()
	deltaBytes := []byte{0x0a, 0x0a, 0x91, 0x00, 0x03, 0x03, 'X', 'Y', 'Z', 0x91, 0x06, 0x04}
	delta := borrowedFromOwned(deltaBytes)
	defer delta.Release()
	out, err := packDeltaApply(base, delta)
	if err != nil {
		t.Fatalf("packDeltaApply error: %v", err)
	}
	if got := string(out.Bytes()); got != "abcXYZghij" {
		t.Fatalf("unexpected delta output: %q", got)
	}
	out.Release()
}

func TestPackDeltaApplyMismatchedBaseSize(t *testing.T) {
	base := borrowedFromOwned([]byte("abc"))
	defer base.Release()
	delta := borrowedFromOwned([]byte{0x04, 0x04})
	defer delta.Release()
	if _, err := packDeltaApply(base, delta); err == nil {
		t.Fatal("expected error for mismatched base size")
	}
}

func TestPackDeltaReadOfsDistance(t *testing.T) {
	dist, err := packDeltaReadOfsDistance(bytes.NewReader([]byte{0x81, 0x01}))
	if err != nil {
		t.Fatalf("packDeltaReadOfsDistance error: %v", err)
	}
	if dist != 257 {
		t.Fatalf("unexpected distance: %d", dist)
	}
	if _, err := packDeltaReadOfsDistance(bytes.NewReader([]byte{})); err == nil {
		t.Fatal("expected error for empty reader")
	}
}

func TestBsearchHash(t *testing.T) {
	h1 := hashWithByte(0x01)
	h2 := hashWithByte(0x03)
	names := append(append([]byte(nil), h1.data[:testHashSize]...), h2.data[:testHashSize]...)
	idx, found := bsearchHash(names, testHashSize, 0, 2, h2)
	if !found || idx != 1 {
		t.Fatalf("expected to find second hash, idx=%d found=%v", idx, found)
	}
	_, found = bsearchHash(names, testHashSize, 0, 2, hashWithByte(0x05))
	if found {
		t.Fatalf("did not expect to find unknown hash")
	}
}

func buildTestPackIndexBuffer(hash Hash, offset uint32) []byte {
	fanout := make([]byte, 256*4)
	first := int(hash.data[0])
	for i := 0; i < 256; i++ {
		var val uint32
		if i >= first {
			val = 1
		}
		binary.BigEndian.PutUint32(fanout[i*4:], val)
	}
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, uint32(idxMagic))
	_ = binary.Write(&buf, binary.BigEndian, uint32(idxVersion2))
	buf.Write(fanout)
	buf.Write(hash.data[:testHashSize])
	buf.Write(make([]byte, 4))
	off32 := make([]byte, 4)
	binary.BigEndian.PutUint32(off32, offset)
	buf.Write(off32)
	buf.Write(make([]byte, 2*testHashSize))
	return buf.Bytes()
}

func TestPackIndexParse(t *testing.T) {
	h := hashWithByte(0x11)
	data := buildTestPackIndexBuffer(h, 0x12345678)
	pi := &packIndex{repo: &Repository{hashSize: testHashSize}}
	if err := pi.parse(data); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if pi.numObjects != 1 {
		t.Fatalf("expected 1 object, got %d", pi.numObjects)
	}
	if got, err := pi.offset(0); err != nil || got != 0x12345678 {
		t.Fatalf("unexpected 32-bit offset or error: %d, %v", got, err)
	}
}

func TestPackIndexOffset64(t *testing.T) {
	pi := &packIndex{}
	pi.offset32 = make([]byte, 4)
	binary.BigEndian.PutUint32(pi.offset32, 0x80000000)
	pi.offset64 = make([]byte, 8)
	binary.BigEndian.PutUint64(pi.offset64, 0x1_0000_0000)
	if got, err := pi.offset(0); err != nil || got != 0x1_0000_0000 {
		t.Fatalf("unexpected 64-bit offset or error: %d, %v", got, err)
	}
}
