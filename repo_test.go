package furgit

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func writeLooseBlob(t *testing.T, root string, data []byte) Hash {
	header, err := headerForType(ObjBlob, data)
	if err != nil {
		t.Fatalf("headerForType: %v", err)
	}
	raw := append(append([]byte(nil), header...), data...)
	id := computeRawHash(raw)
	path := filepath.Join(root, loosePath(id))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for loose object: %v", err)
	}
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	if _, err := zw.Write(raw); err != nil {
		t.Fatalf("compress: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zlib: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write loose object: %v", err)
	}
	return id
}

func TestOpenRepositoryAndLooseRead(t *testing.T) {
	root := t.TempDir()
	repo, err := OpenRepository(root)
	if err != nil {
		t.Fatalf("OpenRepository error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	id := writeLooseBlob(t, root, []byte("loose blob payload"))
	obj, err := repo.looseRead(id)
	if err != nil {
		t.Fatalf("looseRead error: %v", err)
	}
	blob, ok := obj.(*Blob)
	if !ok {
		t.Fatalf("expected Blob, got %T", obj)
	}
	if string(blob.Data) != "loose blob payload" {
		t.Fatalf("blob data mismatch: %q", blob.Data)
	}
}

func TestResolveRefLooseAndPacked(t *testing.T) {
	root := t.TempDir()
	repo, err := OpenRepository(root)
	if err != nil {
		t.Fatalf("OpenRepository error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	looseID := hashWithByte(0xa0)
	loosePath := filepath.Join(root, "refs", "heads")
	if err := os.MkdirAll(loosePath, 0o755); err != nil {
		t.Fatalf("mkdir refs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(loosePath, "master"), []byte(looseID.String()+"\n"), 0o644); err != nil {
		t.Fatalf("write ref: %v", err)
	}
	id, err := repo.ResolveRef("refs/heads/master")
	if err != nil || id != looseID {
		t.Fatalf("ResolveRef loose mismatch (id=%v err=%v)", id, err)
	}

	packedID := hashWithByte(0xb0)
	packed := fmt.Sprintf("%s refs/tags/v1\n", packedID.String())
	if err := os.WriteFile(filepath.Join(root, "packed-refs"), []byte(packed), 0o644); err != nil {
		t.Fatalf("write packed refs: %v", err)
	}
	id, err = repo.resolvePackedRef("refs/tags/v1")
	if err != nil || id != packedID {
		t.Fatalf("resolvePackedRef direct mismatch (id=%v err=%v)", id, err)
	}
	id, err = repo.ResolveRef("refs/tags/v1")
	if err != nil || id != packedID {
		t.Fatalf("ResolveRef packed mismatch (id=%v err=%v)", id, err)
	}

	if _, err := repo.ResolveRef("refs/heads/missing"); !errors.Is(err, ErrInvalidObject) {
		t.Fatalf("expected ErrInvalidObject for missing ref, got %v", err)
	}
}

func TestResolveHEAD(t *testing.T) {
	root := t.TempDir()
	repo, err := OpenRepository(root)
	if err != nil {
		t.Fatalf("OpenRepository error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	headPath := filepath.Join(root, "HEAD")
	if err := os.WriteFile(headPath, []byte("ref: refs/heads/master\n"), 0o644); err != nil {
		t.Fatalf("write HEAD: %v", err)
	}
	ref, err := repo.ResolveHEAD()
	if err != nil || ref != "refs/heads/master" {
		t.Fatalf("ResolveHEAD mismatch (ref=%q err=%v)", ref, err)
	}
	if err := os.WriteFile(headPath, []byte("detached\n"), 0o644); err != nil {
		t.Fatalf("write HEAD detached: %v", err)
	}
	if _, err := repo.ResolveHEAD(); err == nil {
		t.Fatal("expected error for detached HEAD")
	}
}

func TestReadObjectTypeSizeLoose(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	repo, err := OpenRepository(root)
	if err != nil {
		t.Fatalf("OpenRepository error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	data := []byte("header-only read")
	id := writeLooseBlob(t, root, data)
	ty, size, err := repo.ReadObjectTypeSize(id)
	if err != nil {
		t.Fatalf("ReadObjectTypeSize loose error: %v", err)
	}
	if ty != ObjBlob || size != int64(len(data)) {
		t.Fatalf("unexpected loose metadata ty=%d size=%d", ty, size)
	}
}

func TestReadObjectTypeSizePackedObjects(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	objs := []testPackObject{
		{finalType: ObjBlob, body: []byte("packed base payload")},
		{
			finalType: ObjBlob,
			body:      []byte("packed delta payload with extra bytes"),
			encoding:  packEncodingOfsDelta,
			baseIndex: 0,
		},
	}
	ids := writeTestPack(t, root, "pack-basic", objs)

	repo, err := OpenRepository(root)
	if err != nil {
		t.Fatalf("OpenRepository error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	ty, size, err := repo.ReadObjectTypeSize(ids[0])
	if err != nil {
		t.Fatalf("ReadObjectTypeSize base error: %v", err)
	}
	if ty != ObjBlob || size != int64(len(objs[0].body)) {
		t.Fatalf("unexpected base metadata ty=%d size=%d", ty, size)
	}

	ty, size, err = repo.ReadObjectTypeSize(ids[1])
	if err != nil {
		t.Fatalf("ReadObjectTypeSize delta error: %v", err)
	}
	if ty != ObjBlob || size != int64(len(objs[1].body)) {
		t.Fatalf("unexpected delta metadata ty=%d size=%d", ty, size)
	}
}

func TestReadObjectTypeSizePackRefDeltaLooseBase(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	looseBody := []byte("loose base for ref delta")
	baseID := writeLooseBlob(t, root, looseBody)

	objs := []testPackObject{
		{
			finalType: ObjBlob,
			body:      []byte("ref delta rewritten body"),
			encoding:  packEncodingRefDelta,
			baseHash:  baseID,
			baseBody:  looseBody,
		},
	}
	ids := writeTestPack(t, root, "pack-ref", objs)

	repo, err := OpenRepository(root)
	if err != nil {
		t.Fatalf("OpenRepository error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	ty, size, err := repo.ReadObjectTypeSize(ids[0])
	if err != nil {
		t.Fatalf("ReadObjectTypeSize ref delta error: %v", err)
	}
	if ty != ObjBlob || size != int64(len(objs[0].body)) {
		t.Fatalf("unexpected ref delta metadata ty=%d size=%d", ty, size)
	}
}

type packObjectEncoding uint8

const (
	packEncodingFull packObjectEncoding = iota
	packEncodingOfsDelta
	packEncodingRefDelta
)

type testPackObject struct {
	finalType ObjType
	body      []byte
	encoding  packObjectEncoding
	baseIndex int
	baseHash  Hash
	baseBody  []byte
}

func writeTestPack(t *testing.T, root, name string, objs []testPackObject) []Hash {
	t.Helper()
	packDir := filepath.Join(root, "objects", "pack")
	err := os.MkdirAll(packDir, 0o750)
	if err != nil {
		t.Fatalf("mkdir pack dir: %v", err)
	}

	var buf bytes.Buffer
	buf.Write([]byte{'P', 'A', 'C', 'K'})
	err = binary.Write(&buf, binary.BigEndian, uint32(packVersion2))
	if err != nil {
		t.Fatalf("write pack version: %v", err)
	}
	objCount := len(objs)
	if objCount > math.MaxUint32 {
		t.Fatalf("too many objects: %d", len(objs))
	}
	count32 := uint32(objCount) //#nosec G115
	err = binary.Write(&buf, binary.BigEndian, count32)
	if err != nil {
		t.Fatalf("write pack count: %v", err)
	}

	offsets := make([]uint64, len(objs))
	ids := make([]Hash, len(objs))

	for i, obj := range objs {
		offset := buf.Len()
		if offset < 0 {
			t.Fatalf("negative buffer length")
		}
		offsets[i] = uint64(offset)
		header, err := headerForType(obj.finalType, obj.body)
		if err != nil {
			t.Fatalf("headerForType: %v", err)
		}
		raw := make([]byte, len(header)+len(obj.body))
		copy(raw, header)
		copy(raw[len(header):], obj.body)
		ids[i] = computeRawHash(raw)

		switch obj.encoding {
		case packEncodingFull:
			buf.Write(encodePackHeader(obj.finalType, len(obj.body)))
			buf.Write(compressBytes(t, obj.body))
		case packEncodingOfsDelta:
			if obj.baseIndex < 0 || obj.baseIndex >= i {
				t.Fatalf("invalid base index %d for ofs delta %d", obj.baseIndex, i)
			}
			buf.Write(encodePackHeader(ObjOfsDelta, len(obj.body)))
			dist := offsets[i] - offsets[obj.baseIndex]
			buf.Write(encodeOfsDistance(dist))
			baseBody := objs[obj.baseIndex].body
			delta := buildInsertOnlyDelta(len(baseBody), obj.body)
			buf.Write(compressBytes(t, delta))
		case packEncodingRefDelta:
			if obj.baseHash == (Hash{}) {
				t.Fatalf("ref delta %d missing base hash", i)
			}
			baseBody := obj.baseBody
			if len(baseBody) == 0 {
				t.Fatalf("ref delta %d missing base body", i)
			}
			buf.Write(encodePackHeader(ObjRefDelta, len(obj.body)))
			buf.Write(obj.baseHash[:])
			delta := buildInsertOnlyDelta(len(baseBody), obj.body)
			buf.Write(compressBytes(t, delta))
		default:
			t.Fatalf("unknown encoding %d", obj.encoding)
		}
	}

	packContent := append([]byte(nil), buf.Bytes()...)
	packChecksum := newHash(packContent)
	buf.Write(packChecksum[:])
	packBytes := buf.Bytes()

	packPath := filepath.Join(packDir, name+".pack")
	err = os.WriteFile(packPath, packBytes, 0o600)
	if err != nil {
		t.Fatalf("write pack file: %v", err)
	}

	writeTestPackIndex(t, packDir, name, ids, offsets, packChecksum)
	return ids
}

func writeTestPackIndex(t *testing.T, packDir, name string, ids []Hash, offsets []uint64, packChecksum [HashSize]byte) {
	t.Helper()
	type idxEntry struct {
		id     Hash
		offset uint64
	}
	entries := make([]idxEntry, len(ids))
	for i := range ids {
		entries[i] = idxEntry{id: ids[i], offset: offsets[i]}
	}
	sort.Slice(entries, func(i, j int) bool {
		return bytes.Compare(entries[i].id[:], entries[j].id[:]) < 0
	})

	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, uint32(idxMagic))
	if err != nil {
		t.Fatalf("write idx magic: %v", err)
	}
	err = binary.Write(&buf, binary.BigEndian, uint32(idxVersion2))
	if err != nil {
		t.Fatalf("write idx version: %v", err)
	}

	var fanout [256]uint32
	for _, entry := range entries {
		first := int(entry.id[0])
		for i := first; i < len(fanout); i++ {
			fanout[i]++
		}
	}
	for _, count := range fanout {
		err = binary.Write(&buf, binary.BigEndian, count)
		if err != nil {
			t.Fatalf("write fanout: %v", err)
		}
	}

	for _, entry := range entries {
		buf.Write(entry.id[:])
	}

	buf.Write(make([]byte, len(entries)*4))

	for _, entry := range entries {
		if entry.offset >= 0x80000000 {
			t.Fatalf("offset too large for 32-bit table")
		}
		var word [4]byte
		binary.BigEndian.PutUint32(word[:], uint32(entry.offset))
		buf.Write(word[:])
	}

	idxData := append([]byte(nil), buf.Bytes()...)
	idxChecksum := newHash(idxData)
	buf.Write(packChecksum[:])
	buf.Write(idxChecksum[:])

	idxPath := filepath.Join(packDir, name+".idx")
	err = os.WriteFile(idxPath, buf.Bytes(), 0o600)
	if err != nil {
		t.Fatalf("write idx file: %v", err)
	}
}

func buildInsertOnlyDelta(srcLen int, dst []byte) []byte {
	var buf bytes.Buffer
	buf.Write(encodeVarint(srcLen))
	buf.Write(encodeVarint(len(dst)))
	remaining := dst
	for len(remaining) > 0 {
		chunk := remaining
		if len(chunk) > 127 {
			chunk = remaining[:127]
		}
		buf.WriteByte(byte(len(chunk)))
		buf.Write(chunk)
		remaining = remaining[len(chunk):]
	}
	return buf.Bytes()
}

func encodeOfsDistance(dist uint64) []byte {
	if dist == 0 {
		return []byte{0}
	}
	var out []byte
	out = append(out, byte(dist&0x7f))
	for dist >>= 7; dist != 0; dist >>= 7 {
		out = append(out, byte(((dist-1)&0x7f)|0x80))
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}
