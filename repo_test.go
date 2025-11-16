package furgit

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func writeLooseBlob(t *testing.T, repo *Repository, data []byte) Hash {
	blob := &Blob{Data: data}
	id, err := repo.WriteLooseObject(blob)
	if err != nil {
		t.Fatalf("WriteLooseObject: %v", err)
	}
	return id
}

func TestOpenRepositoryAndLooseRead(t *testing.T) {
	root := t.TempDir()
	setupRepoConfig(t, root)
	repo, err := OpenRepository(root)
	if err != nil {
		t.Fatalf("OpenRepository error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	id := writeLooseBlob(t, repo, []byte("loose blob payload"))
	obj, err := repo.looseRead(id)
	if err != nil {
		t.Fatalf("looseRead error: %v", err)
	}
	blob, ok := obj.(*StoredBlob)
	if !ok {
		t.Fatalf("expected StoredBlob, got %T", obj)
	}
	if string(blob.Data) != "loose blob payload" {
		t.Fatalf("blob data mismatch: %q", blob.Data)
	}
}

func TestResolveRefLooseAndPacked(t *testing.T) {
	root := t.TempDir()
	setupRepoConfig(t, root)
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
	setupRepoConfig(t, root)
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
	setupRepoConfig(t, root)
	repo, err := OpenRepository(root)
	if err != nil {
		t.Fatalf("OpenRepository error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	data := []byte("header-only read")
	id := writeLooseBlob(t, repo, data)
	ty, size, err := repo.ReadObjectTypeSize(id)
	if err != nil {
		t.Fatalf("ReadObjectTypeSize loose error: %v", err)
	}
	if ty != ObjectTypeBlob || size != int64(len(data)) {
		t.Fatalf("unexpected loose metadata ty=%d size=%d", ty, size)
	}
}

func TestReadObjectTypeSizePackedObjects(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	setupRepoConfig(t, root)

	objs := []testPackObject{
		{finalType: ObjectTypeBlob, body: []byte("packed base payload")},
		{
			finalType: ObjectTypeBlob,
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
	if ty != ObjectTypeBlob || size != int64(len(objs[0].body)) {
		t.Fatalf("unexpected base metadata ty=%d size=%d", ty, size)
	}

	ty, size, err = repo.ReadObjectTypeSize(ids[1])
	if err != nil {
		t.Fatalf("ReadObjectTypeSize delta error: %v", err)
	}
	if ty != ObjectTypeBlob || size != int64(len(objs[1].body)) {
		t.Fatalf("unexpected delta metadata ty=%d size=%d", ty, size)
	}
}

func TestReadObjectTypeSizePackRefDeltaLooseBase(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	setupRepoConfig(t, root)

	repo, err := OpenRepository(root)
	if err != nil {
		t.Fatalf("OpenRepository error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	looseBody := []byte("loose base for ref delta")
	baseID := writeLooseBlob(t, repo, looseBody)

	objs := []testPackObject{
		{
			finalType: ObjectTypeBlob,
			body:      []byte("ref delta rewritten body"),
			encoding:  packEncodingRefDelta,
			baseHash:  baseID,
			baseBody:  looseBody,
		},
	}
	ids := writeTestPack(t, root, "pack-ref", objs)

	ty, size, err := repo.ReadObjectTypeSize(ids[0])
	if err != nil {
		t.Fatalf("ReadObjectTypeSize ref delta error: %v", err)
	}
	if ty != ObjectTypeBlob || size != int64(len(objs[0].body)) {
		t.Fatalf("unexpected ref delta metadata ty=%d size=%d", ty, size)
	}
}

func TestWriteLooseObjectAllTypes(t *testing.T) {
	root := t.TempDir()
	setupRepoConfig(t, root)
	repo, err := OpenRepository(root)
	if err != nil {
		t.Fatalf("OpenRepository error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	// Blob
	blob := &Blob{Data: []byte("test blob data")}
	blobID, err := repo.WriteLooseObject(blob)
	if err != nil {
		t.Fatalf("WriteLooseObject Blob error: %v", err)
	}
	readBlob, err := repo.ReadObject(blobID)
	if err != nil {
		t.Fatalf("ReadObject Blob error: %v", err)
	}
	if rb, ok := readBlob.(*StoredBlob); !ok {
		t.Fatalf("expected StoredBlob, got %T", readBlob)
	} else if string(rb.Data) != "test blob data" {
		t.Fatalf("blob data mismatch: %q", rb.Data)
	}

	// Tree
	tree := &Tree{
		Entries: []TreeEntry{
			{Mode: 0100644, Name: []byte("file.txt"), ID: blobID},
		},
	}
	treeID, err := repo.WriteLooseObject(tree)
	if err != nil {
		t.Fatalf("WriteLooseObject Tree error: %v", err)
	}
	readTree, err := repo.ReadObject(treeID)
	if err != nil {
		t.Fatalf("ReadObject Tree error: %v", err)
	}
	if rt, ok := readTree.(*StoredTree); !ok {
		t.Fatalf("expected StoredTree, got %T", readTree)
	} else if len(rt.Entries) != 1 {
		t.Fatalf("tree entries mismatch: %d", len(rt.Entries))
	}

	// Commit
	commit := &Commit{
		Tree: treeID,
		Author: Ident{
			Name:          []byte("Test Author"),
			Email:         []byte("test@example.com"),
			WhenUnix:      1700000000,
			OffsetMinutes: 0,
		},
		Committer: Ident{
			Name:          []byte("Test Author"),
			Email:         []byte("test@example.com"),
			WhenUnix:      1700000000,
			OffsetMinutes: 0,
		},
		Message: []byte("Test commit message\n"),
	}
	commitID, err := repo.WriteLooseObject(commit)
	if err != nil {
		t.Fatalf("WriteLooseObject Commit error: %v", err)
	}
	readCommit, err := repo.ReadObject(commitID)
	if err != nil {
		t.Fatalf("ReadObject Commit error: %v", err)
	}
	if rc, ok := readCommit.(*StoredCommit); !ok {
		t.Fatalf("expected StoredCommit, got %T", readCommit)
	} else if rc.Tree != treeID {
		t.Fatalf("commit tree mismatch")
	}

	// Tag
	tag := &Tag{
		Target:     commitID,
		TargetType: ObjectTypeCommit,
		Name:       []byte("v1.0.0"),
		Tagger: &Ident{
			Name:          []byte("Test Tagger"),
			Email:         []byte("tagger@example.com"),
			WhenUnix:      1700000000,
			OffsetMinutes: 0,
		},
		Message: []byte("Test tag message\n"),
	}
	tagID, err := repo.WriteLooseObject(tag)
	if err != nil {
		t.Fatalf("WriteLooseObject Tag error: %v", err)
	}
	readTag, err := repo.ReadObject(tagID)
	if err != nil {
		t.Fatalf("ReadObject Tag error: %v", err)
	}
	if rtag, ok := readTag.(*StoredTag); !ok {
		t.Fatalf("expected StoredTag, got %T", readTag)
	} else if rtag.Target != commitID {
		t.Fatalf("tag target mismatch")
	}
}

type packObjectEncoding uint8

const (
	packEncodingFull packObjectEncoding = iota
	packEncodingOfsDelta
	packEncodingRefDelta
)

type testPackObject struct {
	finalType ObjectType
	body      []byte
	encoding  packObjectEncoding
	baseIndex int
	baseHash  Hash
	baseBody  []byte
}

func writeTestPack(t *testing.T, root, name string, objs []testPackObject) []Hash {
	t.Helper()
	repo := &Repository{hashSize: testHashSize}
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
		ids[i] = repo.computeRawHash(raw)

		switch obj.encoding {
		case packEncodingFull:
			buf.Write(encodePackHeader(obj.finalType, len(obj.body)))
			buf.Write(compressBytes(t, obj.body))
		case packEncodingOfsDelta:
			if obj.baseIndex < 0 || obj.baseIndex >= i {
				t.Fatalf("invalid base index %d for ofs delta %d", obj.baseIndex, i)
			}
			buf.Write(encodePackHeader(ObjectTypeOfsDelta, len(obj.body)))
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
			buf.Write(encodePackHeader(ObjectTypeRefDelta, len(obj.body)))
			buf.Write(obj.baseHash.data[:testHashSize])
			delta := buildInsertOnlyDelta(len(baseBody), obj.body)
			buf.Write(compressBytes(t, delta))
		default:
			t.Fatalf("unknown encoding %d", obj.encoding)
		}
	}

	packContent := append([]byte(nil), buf.Bytes()...)
	packChecksum := repo.computeRawHash(packContent)
	buf.Write(packChecksum.data[:testHashSize])
	packBytes := buf.Bytes()

	packPath := filepath.Join(packDir, name+".pack")
	err = os.WriteFile(packPath, packBytes, 0o600)
	if err != nil {
		t.Fatalf("write pack file: %v", err)
	}

	writeTestPackIndex(t, packDir, name, ids, offsets, packChecksum)
	return ids
}

func writeTestPackIndex(t *testing.T, packDir, name string, ids []Hash, offsets []uint64, packChecksum Hash) {
	t.Helper()
	repo := &Repository{hashSize: testHashSize}
	type idxEntry struct {
		id     Hash
		offset uint64
	}
	entries := make([]idxEntry, len(ids))
	for i := range ids {
		entries[i] = idxEntry{id: ids[i], offset: offsets[i]}
	}
	sort.Slice(entries, func(i, j int) bool {
		return bytes.Compare(entries[i].id.data[:testHashSize], entries[j].id.data[:testHashSize]) < 0
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
		first := int(entry.id.data[0])
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
		buf.Write(entry.id.data[:testHashSize])
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
	idxChecksum := repo.computeRawHash(idxData)
	buf.Write(packChecksum.data[:testHashSize])
	buf.Write(idxChecksum.data[:testHashSize])

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

func setupRepoConfig(t *testing.T, root string) {
	var algo string
	switch testHashSize {
	case sha1.Size:
		algo = "sha1"
	case sha256.Size:
		algo = "sha256"
	default:
		t.Fatalf("unsupported testHashSize: %d", testHashSize)
	}

	cfg := fmt.Sprintf(`
[core]
	repositoryformatversion = 0
[extensions]
	objectformat = %s
`, algo)

	err := os.WriteFile(filepath.Join(root, "config"), []byte(cfg), 0o644)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}
}
