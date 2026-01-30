package furgit

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTagWrite(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	err := os.WriteFile(filepath.Join(workDir, "file.txt"), []byte("content"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file.txt: %v", err)
	}
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "commit", "-m", "Tagged commit")
	commitHash := gitCmd(t, repoPath, nil, "rev-parse", "HEAD")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	commitHashObj, _ := repo.ParseHash(commitHash)

	whenUnix := time.Now().Unix()
	tag := &Tag{
		Target:     commitHashObj,
		TargetType: ObjectTypeCommit,
		Name:       []byte("v2.0.0"),
		Tagger: &Ident{
			Name:          []byte("Tagger Name"),
			Email:         []byte("tagger@test.org"),
			WhenUnix:      whenUnix,
			OffsetMinutes: 120,
		},
		Message: []byte("Release version 2.0.0\n"),
	}

	tagHash, err := repo.WriteLooseObject(tag)
	if err != nil {
		t.Fatalf("WriteLooseObject failed: %v", err)
	}

	gitType := string(gitCatFile(t, repoPath, "-t", tagHash.String()))
	if gitType != "tag" {
		t.Errorf("git type: got %q, want %q", gitType, "tag")
	}

	readObj, err := repo.ReadObject(tagHash)
	if err != nil {
		t.Fatalf("ReadObject failed after write: %v", err)
	}
	readTag, ok := readObj.(*StoredTag)
	if !ok {
		t.Fatalf("expected *StoredTag, got %T", readObj)
	}

	if !bytes.Equal(readTag.Name, []byte("v2.0.0")) {
		t.Errorf("tag name: got %q, want %q", readTag.Name, "v2.0.0")
	}
	if !bytes.HasPrefix(readTag.Tagger.Name, []byte("Tagger Name")) {
		t.Errorf("tagger name: got %q, want prefix %q", readTag.Tagger.Name, "Tagger Name")
	}
	if !bytes.Equal(readTag.Message, []byte("Release version 2.0.0\n")) {
		t.Errorf("message: got %q, want %q", readTag.Message, "Release version 2.0.0\n")
	}

	if tag.ObjectType() != ObjectTypeTag {
		t.Errorf("ObjectType(): got %d, want %d", tag.ObjectType(), ObjectTypeTag)
	}
}

func TestTagRead(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	err := os.WriteFile(filepath.Join(workDir, "file.txt"), []byte("content"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file.txt: %v", err)
	}
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "commit", "-m", "Commit for tag")
	commitHash := gitCmd(t, repoPath, nil, "rev-parse", "HEAD")

	gitCmd(t, repoPath, nil, "tag", "-a", "-m", "Tag message", "v1.0.0", commitHash)
	tagHash := gitCmd(t, repoPath, nil, "rev-parse", "v1.0.0")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	hash, _ := repo.ParseHash(tagHash)
	obj, err := repo.ReadObject(hash)
	if err != nil {
		t.Fatalf("ReadObject failed: %v", err)
	}

	tag, ok := obj.(*StoredTag)
	if !ok {
		t.Fatalf("expected *StoredTag, got %T", obj)
	}

	if !bytes.Equal(tag.Name, []byte("v1.0.0")) {
		t.Errorf("name: got %q, want %q", tag.Name, "v1.0.0")
	}
	if tag.TargetType != ObjectTypeCommit {
		t.Errorf("target type: got %d, want %d", tag.TargetType, ObjectTypeCommit)
	}
	if tag.Target.String() != commitHash {
		t.Errorf("target: got %s, want %s", tag.Target, commitHash)
	}
}

func TestTagRoundtrip(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	err := os.WriteFile(filepath.Join(workDir, "file.txt"), []byte("content"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file.txt: %v", err)
	}
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "commit", "-m", "Commit")
	commitHash := gitCmd(t, repoPath, nil, "rev-parse", "HEAD")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	commitHashObj, _ := repo.ParseHash(commitHash)

	tag := &Tag{
		Target:     commitHashObj,
		TargetType: ObjectTypeCommit,
		Name:       []byte("v3.0.0"),
		Tagger: &Ident{
			Name:          []byte("Test Tagger"),
			Email:         []byte("tagger@example.org"),
			WhenUnix:      123456789,
			OffsetMinutes: 0,
		},
		Message: []byte("Tag message\n"),
	}

	tagHash, err := repo.WriteLooseObject(tag)
	if err != nil {
		t.Fatalf("WriteLooseObject failed: %v", err)
	}

	obj, err := repo.ReadObject(tagHash)
	if err != nil {
		t.Fatalf("ReadObject failed: %v", err)
	}

	readTag, ok := obj.(*StoredTag)
	if !ok {
		t.Fatalf("expected *StoredTag, got %T", obj)
	}

	if !bytes.Equal(readTag.Name, tag.Name) {
		t.Errorf("name: got %q, want %q", readTag.Name, tag.Name)
	}
	if readTag.Target != tag.Target {
		t.Errorf("target: got %s, want %s", readTag.Target, tag.Target)
	}
	if readTag.TargetType != tag.TargetType {
		t.Errorf("target type: got %d, want %d", readTag.TargetType, tag.TargetType)
	}
	if !bytes.Equal(readTag.Message, tag.Message) {
		t.Errorf("message: got %q, want %q", readTag.Message, tag.Message)
	}
}
