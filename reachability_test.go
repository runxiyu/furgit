package furgit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReachabilityCommitsWantHave(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	var commits []string
	for i := 0; i < 3; i++ {
		path := filepath.Join(workDir, "file.txt")
		if err := os.WriteFile(path, []byte{byte('a' + i), '\n'}, 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}
		gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
		gitCmd(t, repoPath, nil, "--work-tree="+workDir, "commit", "-m", "commit")
		commits = append(commits, gitCmd(t, repoPath, nil, "rev-parse", "HEAD"))
	}

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	wantID, _ := repo.ParseHash(commits[2])
	haveID, _ := repo.ParseHash(commits[1])
	walk, err := repo.ReachableObjects(ReachabilityQuery{
		Wants: []Hash{wantID},
		Haves: []Hash{haveID},
		Mode:  ReachabilityCommitsOnly,
	})
	if err != nil {
		t.Fatalf("ReachableObjects failed: %v", err)
	}

	seen := make(map[Hash]ReachableObject)
	for obj := range walk.Seq() {
		seen[obj.ID] = obj
		if obj.Type != ObjectTypeCommit {
			t.Fatalf("unexpected object type: %v", obj.Type)
		}
	}
	if err := walk.Err(); err != nil {
		t.Fatalf("Reachability walk error: %v", err)
	}

	headID := wantID
	parentID, _ := repo.ParseHash(commits[1])
	rootID, _ := repo.ParseHash(commits[0])
	if _, ok := seen[headID]; !ok {
		t.Fatalf("missing head commit")
	}
	if _, ok := seen[parentID]; !ok {
		t.Fatalf("missing parent commit")
	}
	if _, ok := seen[rootID]; !ok {
		t.Fatalf("missing root commit")
	}
	if seen[headID].InHave {
		t.Fatalf("head commit incorrectly marked InHave")
	}
	if !seen[parentID].InHave || !seen[rootID].InHave {
		t.Fatalf("expected parent and root commits to be InHave")
	}

	inHave, err := walk.HaveContains(parentID)
	if err != nil {
		t.Fatalf("HaveContains failed: %v", err)
	}
	if !inHave {
		t.Fatalf("expected parent to be reachable from have")
	}
}

func TestReachabilityAllObjects(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	if err := os.WriteFile(filepath.Join(workDir, "file1.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatalf("write file1: %v", err)
	}
	if err := os.Mkdir(filepath.Join(workDir, "dir"), 0o755); err != nil {
		t.Fatalf("mkdir dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "dir", "file2.txt"), []byte("two\n"), 0o644); err != nil {
		t.Fatalf("write file2: %v", err)
	}
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
	gitCmd(t, repoPath, nil, "--work-tree="+workDir, "commit", "-m", "commit")

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	head := gitCmd(t, repoPath, nil, "rev-parse", "HEAD")
	wantID, _ := repo.ParseHash(head)
	walk, err := repo.ReachableObjects(ReachabilityQuery{
		Wants: []Hash{wantID},
		Mode:  ReachabilityAllObjects,
	})
	if err != nil {
		t.Fatalf("ReachableObjects failed: %v", err)
	}

	seen := make(map[Hash]ObjectType)
	for obj := range walk.Seq() {
		seen[obj.ID] = obj.Type
	}
	if err := walk.Err(); err != nil {
		t.Fatalf("Reachability walk error: %v", err)
	}

	treeStr := gitCmd(t, repoPath, nil, "show", "-s", "--format=%T", head)
	treeID, _ := repo.ParseHash(treeStr)
	lsTree := gitCmd(t, repoPath, nil, "ls-tree", "-r", treeStr)
	fields := strings.Fields(lsTree)
	if len(fields) < 3 {
		t.Fatalf("unexpected ls-tree output: %q", lsTree)
	}
	blobID, _ := repo.ParseHash(fields[2])

	if seen[wantID] != ObjectTypeCommit {
		t.Fatalf("missing commit in reachability walk")
	}
	if seen[treeID] != ObjectTypeTree {
		t.Fatalf("missing tree in reachability walk")
	}
	if seen[blobID] != ObjectTypeBlob {
		t.Fatalf("missing blob in reachability walk")
	}
}

func TestReachabilityStopAtHaves(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	workDir, cleanupWork := setupWorkDir(t)
	defer cleanupWork()

	var commits []string
	for i := 0; i < 3; i++ {
		path := filepath.Join(workDir, "file.txt")
		if err := os.WriteFile(path, []byte{byte('a' + i), '\n'}, 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}
		gitCmd(t, repoPath, nil, "--work-tree="+workDir, "add", ".")
		gitCmd(t, repoPath, nil, "--work-tree="+workDir, "commit", "-m", "commit")
		commits = append(commits, gitCmd(t, repoPath, nil, "rev-parse", "HEAD"))
	}

	repo, err := OpenRepository(repoPath)
	if err != nil {
		t.Fatalf("OpenRepository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	wantID, _ := repo.ParseHash(commits[2])
	haveID, _ := repo.ParseHash(commits[1])
	walk, err := repo.ReachableObjects(ReachabilityQuery{
		Wants:       []Hash{wantID},
		Haves:       []Hash{haveID},
		Mode:        ReachabilityCommitsOnly,
		StopAtHaves: true,
	})
	if err != nil {
		t.Fatalf("ReachableObjects failed: %v", err)
	}

	var got []Hash
	for obj := range walk.Seq() {
		got = append(got, obj.ID)
		if obj.InHave {
			t.Fatalf("unexpected InHave object in send set")
		}
	}
	if err := walk.Err(); err != nil {
		t.Fatalf("Reachability walk error: %v", err)
	}
	if len(got) != 1 || got[0] != wantID {
		t.Fatalf("StopAtHaves mismatch: got %d objects", len(got))
	}
}
