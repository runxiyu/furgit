package repository_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/repository"
)

func TestRepositoryDepthFirstEnumerationFromHEAD(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repoHarness := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		_, _, commit1 := repoHarness.MakeCommit(t, "walk-one")
		blob2, tree2 := repoHarness.MakeSingleFileTree(t, "second.txt", []byte("second\n"))
		commit2 := repoHarness.CommitTree(t, tree2, "walk-two", commit1)
		_ = blob2

		repoHarness.UpdateRef(t, "refs/heads/main", commit2)
		repoHarness.SymbolicRef(t, "HEAD", "refs/heads/main")

		walkRepositoryFromHead(t, repoHarness.Dir())
	})
}

func TestRepositoryDepthFirstEnumerationCurrentWorktree(t *testing.T) {
	t.Parallel()

	worktreeRoot := filepath.Clean("..")
	gitPath := filepath.Join(worktreeRoot, ".git")

	info, err := os.Stat(gitPath)
	if err != nil {
		t.Fatalf("stat %q: %v", gitPath, err)
	}

	if info.IsDir() {
		walkRepositoryFromHead(t, gitPath)

		return
	}

	if !info.Mode().IsRegular() {
		t.Fatalf("%q is neither a directory nor a regular file", gitPath)
	}

	content, err := os.ReadFile(gitPath) //#nosec G304
	if err != nil {
		t.Fatalf("read %q: %v", gitPath, err)
	}

	line := strings.TrimSpace(string(content))

	prefix := "gitdir: "
	if !strings.HasPrefix(line, prefix) {
		t.Fatalf("%q file does not begin with %q", gitPath, prefix)
	}

	gitdirRel := strings.TrimSpace(line[len(prefix):])
	if gitdirRel == "" {
		t.Fatalf("%q contains empty gitdir path", gitPath)
	}

	gitdirPath := gitdirRel
	if !filepath.IsAbs(gitdirPath) {
		gitdirPath = filepath.Join(worktreeRoot, gitdirPath)
	}

	commondirPath := filepath.Join(gitdirPath, "commondir")

	commondirContent, err := os.ReadFile(commondirPath) //#nosec G304
	if err != nil {
		t.Fatalf("read %q: %v", commondirPath, err)
	}

	repoPath := strings.TrimSpace(string(commondirContent))
	if repoPath == "" {
		t.Fatalf("%q contains empty repo path", commondirPath)
	}

	if filepath.IsAbs(repoPath) {
		walkRepositoryFromHead(t, repoPath)

		return
	}

	repoPath = filepath.Join(gitdirPath, repoPath)

	walkRepositoryFromHead(t, repoPath)
}

func walkRepositoryFromHead(t *testing.T, repoPath string) {
	t.Helper()

	root, err := os.OpenRoot(repoPath)
	if err != nil {
		t.Fatalf("os.OpenRoot(%q): %v", repoPath, err)
	}

	defer func() { _ = root.Close() }()

	repo, err := repository.Open(root)
	if err != nil {
		t.Fatalf("repository.Open(root for %q): %v", repoPath, err)
	}

	defer func() { _ = repo.Close() }()

	head, err := repo.Refs().ResolveFully("HEAD")
	if err != nil {
		t.Fatalf("ResolveRefFully(HEAD): %v", err)
	}

	objectsRead, err := traverseReachableIter(repo, head.ID)
	if err != nil {
		t.Fatalf("traverseReachableIter(%s): %v", head.ID, err)
	}

	if objectsRead <= 0 {
		t.Fatalf("no objects were enumerated from HEAD (%s)", fmt.Sprintf("%q", repoPath))
	}
}

func traverseReachableIter(repo *repository.Repository, root objectid.ObjectID) (int, error) {
	stack := []objectid.ObjectID{root}
	visited := make(map[objectid.ObjectID]struct{})
	total := 0

	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		_, ok := visited[id]
		if ok {
			continue
		}

		visited[id] = struct{}{}

		stored, err := repo.ReadStored(id)
		if err != nil {
			return 0, err
		}

		total++

		switch obj := stored.Object().(type) {
		case *object.Commit:
			stack = append(stack, obj.Tree)
			stack = append(stack, obj.Parents...)
		case *object.Tree:
			for i := len(obj.Entries) - 1; i >= 0; i-- {
				entry := obj.Entries[i]
				if entry.Mode == object.FileModeGitlink {
					continue
				}

				stack = append(stack, entry.ID)
			}
		case *object.Tag:
			stack = append(stack, obj.Target)
		case *object.Blob:
		default:
			// Unknown parsed object variants are treated as leaves.
		}
	}

	return total, nil
}
