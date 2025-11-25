package furgit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveRef(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	err := os.WriteFile(filepath.Join(workDir, "test.txt"), []byte("content"), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test.txt: %v", err)
	}
	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "test")
	commitHash := gitCmd(t, repoPath, "rev-parse", "HEAD")
	gitCmd(t, repoPath, "update-ref", "refs/heads/main", commitHash)

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	hashObj, _ := repo.ParseHash(commitHash)
	resolved, err := repo.ResolveRef("refs/heads/main")
	if err != nil {
		t.Fatalf("ResolveRef failed: %v", err)
	}

	if resolved.Kind != RefKindDetached {
		t.Fatalf("expected detached ref, got %v", resolved.Kind)
	}
	if resolved.Hash != hashObj {
		t.Errorf("resolved hash: got %s, want %s", resolved.Hash, hashObj)
	}

	gitRevParse := gitCmd(t, repoPath, "rev-parse", "refs/heads/main")
	if resolved.Hash.String() != gitRevParse {
		t.Errorf("furgit resolved %s, git resolved %s", resolved.Hash, gitRevParse)
	}

	_, err = repo.ResolveRef("refs/heads/nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent ref")
	}
}

func TestResolveHEAD(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	err := os.WriteFile(filepath.Join(workDir, "test.txt"), []byte("content"), 0o644)
	if err != nil {
		t.Fatalf("failed to write test.txt: %v", err)
	}
	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "test")
	commitHash := gitCmd(t, repoPath, "rev-parse", "HEAD")
	gitCmd(t, repoPath, "update-ref", "refs/heads/main", commitHash)
	gitCmd(t, repoPath, "symbolic-ref", "HEAD", "refs/heads/main")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	ref, err := repo.ResolveRef("HEAD")
	if err != nil {
		t.Fatalf("ResolveRef(HEAD) failed: %v", err)
	}

	if ref.Kind != RefKindSymbolic {
		t.Fatalf("HEAD kind: got %v, want %v", ref.Kind, RefKindSymbolic)
	}

	if ref.Ref != "refs/heads/main" {
		t.Errorf("HEAD symbolic ref: got %q, want %q", ref.Ref, "refs/heads/main")
	}

	gitSymRef := gitCmd(t, repoPath, "symbolic-ref", "HEAD")
	if ref.Ref != gitSymRef {
		t.Errorf("furgit resolved %v, git resolved %s", ref.Ref, gitSymRef)
	}
}

func TestPackedRefs(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	err := os.WriteFile(filepath.Join(workDir, "test.txt"), []byte("content1"), 0o644)
	if err != nil {
		t.Fatalf("failed to write test.txt: %v", err)
	}
	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "commit1")
	commit1Hash := gitCmd(t, repoPath, "rev-parse", "HEAD")

	err = os.WriteFile(filepath.Join(workDir, "test2.txt"), []byte("content2"), 0o644)
	if err != nil {
		t.Fatalf("failed to write test2.txt: %v", err)
	}
	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "commit2")
	commit2Hash := gitCmd(t, repoPath, "rev-parse", "HEAD")

	gitCmd(t, repoPath, "update-ref", "refs/heads/branch1", commit1Hash)
	gitCmd(t, repoPath, "update-ref", "refs/heads/branch2", commit2Hash)
	gitCmd(t, repoPath, "update-ref", "refs/tags/v1.0", commit1Hash)

	gitCmd(t, repoPath, "pack-refs", "--all")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	hash1, _ := repo.ParseHash(commit1Hash)
	hash2, _ := repo.ParseHash(commit2Hash)

	resolved1, err := repo.ResolveRef("refs/heads/branch1")
	if err != nil {
		t.Fatalf("ResolveRef branch1 failed: %v", err)
	}
	if resolved1.Kind != RefKindDetached || resolved1.Hash != hash1 {
		t.Errorf("branch1: got %s, want %s", resolved1.Hash, hash1)
	}

	gitResolved1 := gitCmd(t, repoPath, "rev-parse", "refs/heads/branch1")
	if resolved1.Hash.String() != gitResolved1 {
		t.Errorf("furgit resolved %s, git resolved %s", resolved1.Hash, gitResolved1)
	}

	resolved2, err := repo.ResolveRef("refs/heads/branch2")
	if err != nil {
		t.Fatalf("ResolveRef branch2 failed: %v", err)
	}
	if resolved2.Kind != RefKindDetached || resolved2.Hash != hash2 {
		t.Errorf("branch2: got %s, want %s", resolved2.Hash, hash2)
	}

	resolvedTag, err := repo.ResolveRef("refs/tags/v1.0")
	if err != nil {
		t.Fatalf("ResolveRef tag failed: %v", err)
	}
	if resolvedTag.Kind != RefKindDetached || resolvedTag.Hash != hash1 {
		t.Errorf("tag: got %s, want %s", resolvedTag.Hash, hash1)
	}
}

func TestResolveRefFully(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	// Create an initial commit
	err := os.WriteFile(filepath.Join(workDir, "file.txt"), []byte("content"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file.txt: %v", err)
	}
	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "init")
	commit := gitCmd(t, repoPath, "rev-parse", "HEAD")

	// Create two layers of symbolic refs
	gitCmd(t, repoPath, "symbolic-ref", "refs/heads/level1", "refs/heads/level2")
	gitCmd(t, repoPath, "symbolic-ref", "refs/heads/level2", "refs/heads/main")
	gitCmd(t, repoPath, "update-ref", "refs/heads/main", commit)

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	commitHash, err := repo.ParseHash(commit)
	if err != nil {
		t.Fatalf("ParseHash failed: %v", err)
	}

	resolved, err := repo.ResolveRefFully("refs/heads/level1")
	if err != nil {
		t.Fatalf("ResolveRefFully failed: %v", err)
	}

	if resolved != commitHash {
		t.Errorf("ResolveRefFully: got hash %s, want %s", resolved, commitHash)
	}
}

func TestResolveRefFullySymbolicCycle(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	gitCmd(t, repoPath, "symbolic-ref", "refs/heads/A", "refs/heads/B")
	gitCmd(t, repoPath, "symbolic-ref", "refs/heads/B", "refs/heads/A")

	_, err = repo.ResolveRefFully("refs/heads/A")
	if err == nil {
		t.Fatalf("ResolveRefFully should fail on a symbolic cycle")
	}

	if !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("unexpected error for symbolic cycle: %v", err)
	}
}

func TestResolveRefHashInput(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	err := os.WriteFile(filepath.Join(workDir, "file.txt"), []byte("content"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file.txt: %v", err)
	}
	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "init")

	commitHash := gitCmd(t, repoPath, "rev-parse", "HEAD")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	hashObj, err := repo.ParseHash(commitHash)
	if err != nil {
		t.Fatalf("ParseHash failed: %v", err)
	}

	ref, err := repo.ResolveRef(commitHash)
	if err != nil {
		t.Fatalf("ResolveRef(hash) failed: %v", err)
	}
	if ref.Kind != RefKindDetached {
		t.Fatalf("expected RefKindDetached, got %v", ref.Kind)
	}
	if ref.Hash != hashObj {
		t.Fatalf("hash mismatch: got %s, want %s", ref.Hash, hashObj)
	}

	hash, err := repo.ResolveRefFully(commitHash)
	if err != nil {
		t.Fatalf("ResolveRefFully(hash) failed: %v", err)
	}
	if hash != hashObj {
		t.Fatalf("hash mismatch: got %s, want %s", hash, hashObj)
	}

	_, err = repo.ResolveRef("this_is_not_a_hash")
	if err == nil {
		t.Fatalf("expected error for invalid hash input")
	}
}

func TestListRefsLooseOverridesPacked(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	gitCmd(t, repoPath, "symbolic-ref", "HEAD", "refs/heads/main")

	err := os.WriteFile(filepath.Join(workDir, "file.txt"), []byte("one"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file.txt: %v", err)
	}
	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "c1")
	commit1 := gitCmd(t, repoPath, "rev-parse", "HEAD")

	gitCmd(t, repoPath, "update-ref", "refs/heads/main", commit1)
	gitCmd(t, repoPath, "update-ref", "refs/heads/feature", commit1)
	gitCmd(t, repoPath, "pack-refs", "--all", "--prune")

	err = os.WriteFile(filepath.Join(workDir, "file.txt"), []byte("two"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file.txt: %v", err)
	}
	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "c2")
	commit2 := gitCmd(t, repoPath, "rev-parse", "HEAD")
	gitCmd(t, repoPath, "update-ref", "refs/heads/main", commit2)

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	hash1, _ := repo.ParseHash(commit1)
	hash2, _ := repo.ParseHash(commit2)

	refs, err := repo.ListRefs("refs/heads/*")
	if err != nil {
		t.Fatalf("ListRefs failed: %v", err)
	}

	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %d", len(refs))
	}

	got := make(map[string]Ref, len(refs))
	for _, r := range refs {
		if _, exists := got[r.Name]; exists {
			t.Fatalf("duplicate ref %q in results", r.Name)
		}
		got[r.Name] = r.Ref
	}

	mainRef, ok := got["refs/heads/main"]
	if !ok {
		t.Fatalf("missing refs/heads/main in results")
	}
	if mainRef.Kind != RefKindDetached || mainRef.Hash != hash2 {
		t.Fatalf("refs/heads/main hash: got %s (kind %v), want %s", mainRef.Hash, mainRef.Kind, hash2)
	}

	featureRef, ok := got["refs/heads/feature"]
	if !ok {
		t.Fatalf("missing refs/heads/feature in results")
	}
	if featureRef.Kind != RefKindDetached || featureRef.Hash != hash1 {
		t.Fatalf("refs/heads/feature hash: got %s (kind %v), want %s", featureRef.Hash, featureRef.Kind, hash1)
	}
}

func TestListRefsPatternFiltering(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	gitCmd(t, repoPath, "symbolic-ref", "HEAD", "refs/heads/main")

	err := os.WriteFile(filepath.Join(workDir, "file.txt"), []byte("one"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file.txt: %v", err)
	}
	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "c1")
	commit1 := gitCmd(t, repoPath, "rev-parse", "HEAD")

	gitCmd(t, repoPath, "update-ref", "refs/heads/main", commit1)
	gitCmd(t, repoPath, "update-ref", "refs/heads/feature", commit1)
	gitCmd(t, repoPath, "pack-refs", "--all", "--prune")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	hash1, _ := repo.ParseHash(commit1)

	refs, err := repo.ListRefs("refs/heads/fea*")
	if err != nil {
		t.Fatalf("ListRefs failed: %v", err)
	}
	if len(refs) != 1 {
		t.Fatalf("expected 1 ref, got %d", len(refs))
	}
	if refs[0].Name != "refs/heads/feature" {
		t.Fatalf("unexpected ref name: got %q, want %q", refs[0].Name, "refs/heads/feature")
	}
	if refs[0].Ref.Kind != RefKindDetached || refs[0].Ref.Hash != hash1 {
		t.Fatalf("refs/heads/feature hash: got %s (kind %v), want %s", refs[0].Ref.Hash, refs[0].Ref.Kind, hash1)
	}
}

func TestListRefsPackedPatterns(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	gitCmd(t, repoPath, "symbolic-ref", "HEAD", "refs/heads/main")

	err := os.WriteFile(filepath.Join(workDir, "file.txt"), []byte("one"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file.txt: %v", err)
	}
	gitCmd(t, repoPath, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, "--work-tree="+workDir, "commit", "-m", "c1")
	commit := gitCmd(t, repoPath, "rev-parse", "HEAD")

	gitCmd(t, repoPath, "update-ref", "refs/heads/main", commit)
	gitCmd(t, repoPath, "update-ref", "refs/heads/feature/one", commit)
	gitCmd(t, repoPath, "update-ref", "refs/notes/review", commit)
	gitCmd(t, repoPath, "update-ref", "refs/tags/v1", commit)
	gitCmd(t, repoPath, "pack-refs", "--all", "--prune")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	tests := []struct {
		pattern string
		want    []string
	}{
		{
			pattern: "refs/heads/*",
			want:    []string{"refs/heads/main"},
		},
		{
			pattern: "refs/heads/*/*",
			want:    []string{"refs/heads/feature/one"},
		},
		{
			pattern: "refs/*/feature/one",
			want:    []string{"refs/heads/feature/one"},
		},
		{
			pattern: "refs/heads/feat?re/one",
			want:    []string{"refs/heads/feature/one"},
		},
		{
			pattern: "refs/tags/v[0-9]",
			want:    []string{"refs/tags/v1"},
		},
		{
			pattern: "refs/*/*",
			want:    []string{"refs/heads/main", "refs/notes/review", "refs/tags/v1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			refs, err := repo.ListRefs(tt.pattern)
			if err != nil {
				t.Fatalf("ListRefs(%q) failed: %v", tt.pattern, err)
			}

			got := make(map[string]struct{}, len(refs))
			for _, r := range refs {
				got[r.Name] = struct{}{}
			}

			want := make(map[string]struct{}, len(tt.want))
			for _, w := range tt.want {
				want[w] = struct{}{}
			}

			if len(got) != len(want) {
				t.Fatalf("ListRefs(%q) returned %d refs, want %d", tt.pattern, len(got), len(want))
			}
			for name := range got {
				if _, ok := want[name]; !ok {
					t.Fatalf("ListRefs(%q) unexpected ref %q", tt.pattern, name)
				}
			}
		})
	}
}
