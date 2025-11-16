package furgit

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func mustHash(t *testing.T, hex string) Hash {
	id, err := ParseHashWithSize(hex, testHashSize)
	if err != nil {
		t.Fatalf("ParseHash failed: %v", err)
	}
	return id
}

func hashWithByte(fill byte) Hash {
	var h Hash
	for i := 0; i < testHashSize; i++ {
		h[i] = fill
		fill++
	}
	return h
}

func TestLoosePathUsesExpectedLayout(t *testing.T) {
	pattern := "0123456789abcdef"
	repeats := (testHashSize*2 + len(pattern) - 1) / len(pattern)
	hexStr := strings.Repeat(pattern, repeats)[:testHashSize*2]
	id := mustHash(t, hexStr)
	expect := filepath.Join("objects", hexStr[:2], hexStr[2:])
	if got := loosePath(id, testHashSize); got != expect {
		t.Fatalf("unexpected loose path: %q", got)
	}
}

func TestParseBlobAndSerialize(t *testing.T) {
	data := []byte("blob payload")
	id := hashWithByte(0x10)
	blob, err := parseBlob(id, data)
	if err != nil {
		t.Fatalf("parseBlob error: %v", err)
	}
	if !bytes.Equal(blob.Data, data) {
		t.Fatalf("blob data mismatch: %q", blob.Data)
	}
	if blob.Hash != id {
		t.Fatalf("blob hash mismatch: %v", blob.Hash)
	}
	raw, err := blob.Serialize(testHashSize)
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	header, err := headerForType(ObjBlob, data)
	if err != nil {
		t.Fatalf("headerForType: %v", err)
	}
	want := append(append([]byte(nil), header...), data...)
	if !bytes.Equal(raw, want) {
		t.Fatalf("serialized blob mismatch")
	}
}

func TestParseTreeAndSerialize(t *testing.T) {
	entries := []TreeEntry{
		{Mode: 0100644, Name: []byte("file.txt"), ID: hashWithByte(0x20)},
		{Mode: 040000, Name: []byte("subdir"), ID: hashWithByte(0x30)},
	}
	body := treeBody(&Tree{Entries: entries}, testHashSize)
	id := hashWithByte(0x40)
	tree, err := parseTree(id, body, testHashSize)
	if err != nil {
		t.Fatalf("parseTree error: %v", err)
	}
	if len(tree.Entries) != len(entries) {
		t.Fatalf("expected %d entries, got %d", len(entries), len(tree.Entries))
	}
	for i := range entries {
		if tree.Entries[i].Mode != entries[i].Mode || !bytes.Equal(tree.Entries[i].Name, entries[i].Name) || tree.Entries[i].ID != entries[i].ID {
			t.Fatalf("entry %d mismatch", i)
		}
	}
	serialized, err := (&Tree{Entries: entries}).Serialize(testHashSize)
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	header, _ := headerForType(ObjTree, body)
	want := append(append([]byte(nil), header...), body...)
	if !bytes.Equal(serialized, want) {
		t.Fatalf("serialized tree mismatch")
	}
}

func TestParseCommitWithExtraHeader(t *testing.T) {
	treeID := hashWithByte(0x50)
	parent := hashWithByte(0x60)
	ident := Ident{
		Name:          []byte("Alice"),
		Email:         []byte("alice@example.com"),
		WhenUnix:      1700000000,
		OffsetMinutes: -420,
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "tree %s\n", treeID.StringWithSize(testHashSize))
	fmt.Fprintf(&buf, "parent %s\n", parent.StringWithSize(testHashSize))
	buf.WriteString("author ")
	buf.Write(ident.Serialize())
	buf.WriteByte('\n')
	buf.WriteString("committer ")
	buf.Write(ident.Serialize())
	buf.WriteByte('\n')
	buf.WriteString("extra data\n\nMessage body\n")
	commit, err := parseCommit(hashWithByte(0x70), buf.Bytes(), testHashSize)
	if err != nil {
		t.Fatalf("parseCommit error: %v", err)
	}
	if commit.Tree != treeID {
		t.Fatalf("tree mismatch")
	}
	if len(commit.Parents) != 1 || commit.Parents[0] != parent {
		t.Fatalf("parent mismatch: %+v", commit.Parents)
	}
	if string(commit.Message) != "Message body\n" {
		t.Fatalf("message mismatch: %q", commit.Message)
	}
	if len(commit.ExtraHeaders) != 1 || commit.ExtraHeaders[0].Key != "extra" || !bytes.Equal(commit.ExtraHeaders[0].Value, []byte("data")) {
		t.Fatalf("extra headers mismatch: %+v", commit.ExtraHeaders)
	}

	roundTrip := &Commit{
		Tree:      treeID,
		Parents:   []Hash{parent},
		Author:    ident,
		Committer: ident,
		Message:   []byte("Message body\n"),
	}
	raw, err := roundTrip.Serialize(testHashSize)
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	if !strings.Contains(string(raw), "tree "+treeID.StringWithSize(testHashSize)) {
		t.Fatalf("serialized commit missing tree header")
	}
}

func TestParseTagAndSerialize(t *testing.T) {
	target := hashWithByte(0x80)
	tagger := &Ident{
		Name:          []byte("Tagger"),
		Email:         []byte("tagger@example.com"),
		WhenUnix:      1234,
		OffsetMinutes: 0,
	}
	var buf bytes.Buffer
	buf.WriteString("object ")
	buf.WriteString(target.StringWithSize(testHashSize))
	buf.WriteByte('\n')
	buf.WriteString("type commit\n")
	buf.WriteString("tag v1.0\n")
	buf.WriteString("tagger ")
	buf.Write(tagger.Serialize())
	buf.WriteString("\n\nannotated tag\n")
	body := append([]byte(nil), buf.Bytes()...)
	tag, err := parseTag(hashWithByte(0x90), body, testHashSize)
	if err != nil {
		t.Fatalf("parseTag error: %v", err)
	}
	if tag.Target != target || tag.TargetType != ObjCommit {
		t.Fatalf("tag target mismatch")
	}
	if tag.Tagger == nil {
		t.Fatalf("tagger missing in body %q", string(body))
	}
	if !bytes.Contains(tag.Tagger.Name, []byte("Tagger")) {
		t.Fatalf("tagger name mismatch: %q", tag.Tagger.Name)
	}
	if string(tag.Name) != "v1.0" {
		t.Fatalf("tag name mismatch: %q", tag.Name)
	}
	serialized, err := tag.Serialize(testHashSize)
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	if !strings.Contains(string(serialized), "tag v1.0") {
		t.Fatalf("serialized tag missing name header")
	}
}
