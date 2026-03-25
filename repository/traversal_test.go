package repository_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	objectid "codeberg.org/lindenii/furgit/object/id"
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

		root := repoHarness.OpenGitRoot(t)
		walkRepositoryFromRoot(t, root, "test repo")
	})
}

func TestRepositoryDepthFirstEnumerationCurrentWorktree(t *testing.T) {
	t.Parallel()

	worktreeRoot := filepath.Clean("..")

	worktreeFS, err := os.OpenRoot(worktreeRoot)
	if err != nil {
		t.Fatalf("os.OpenRoot(%q): %v", worktreeRoot, err)
	}

	defer func() { _ = worktreeFS.Close() }()

	info, err := worktreeFS.Stat(".git")
	if err != nil {
		t.Fatalf("stat %q: %v", filepath.Join(worktreeRoot, ".git"), err)
	}

	if info.IsDir() {
		gitRoot, err := worktreeFS.OpenRoot(".git")
		if err != nil {
			t.Fatalf("OpenRoot(.git): %v", err)
		}

		defer func() { _ = gitRoot.Close() }()

		walkRepositoryFromRoot(t, gitRoot, filepath.Join(worktreeRoot, ".git"))

		return
	}

	if !info.Mode().IsRegular() {
		t.Fatalf("%q is neither a directory nor a regular file", filepath.Join(worktreeRoot, ".git"))
	}

	content, err := worktreeFS.ReadFile(".git")
	if err != nil {
		t.Fatalf("read %q: %v", filepath.Join(worktreeRoot, ".git"), err)
	}

	line := strings.TrimSpace(string(content))

	prefix := "gitdir: "
	if !strings.HasPrefix(line, prefix) {
		t.Fatalf("%q file does not begin with %q", filepath.Join(worktreeRoot, ".git"), prefix)
	}

	gitdirRel := strings.TrimSpace(line[len(prefix):])
	if gitdirRel == "" {
		t.Fatalf("%q contains empty gitdir path", filepath.Join(worktreeRoot, ".git"))
	}

	gitdirPath := gitdirRel
	if !filepath.IsAbs(gitdirPath) {
		gitdirPath = filepath.Join(worktreeRoot, gitdirPath)
	}

	gitRoot, err := os.OpenRoot(gitdirPath)
	if err != nil {
		t.Fatalf("os.OpenRoot(%q): %v", gitdirPath, err)
	}

	defer func() { _ = gitRoot.Close() }()

	commondirContent, err := gitRoot.ReadFile("commondir")
	if err != nil {
		t.Fatalf("read %q: %v", filepath.Join(gitdirPath, "commondir"), err)
	}

	repoPath := strings.TrimSpace(string(commondirContent))
	if repoPath == "" {
		t.Fatalf("%q contains empty repo path", filepath.Join(gitdirPath, "commondir"))
	}

	if filepath.IsAbs(repoPath) {
		repoRoot, err := os.OpenRoot(repoPath)
		if err != nil {
			t.Fatalf("os.OpenRoot(%q): %v", repoPath, err)
		}

		defer func() { _ = repoRoot.Close() }()

		walkRepositoryFromRoot(t, repoRoot, repoPath)

		return
	}

	repoPath = filepath.Join(gitdirPath, repoPath)

	repoRoot, err := os.OpenRoot(repoPath)
	if err != nil {
		t.Fatalf("os.OpenRoot(%q): %v", repoPath, err)
	}

	defer func() { _ = repoRoot.Close() }()

	walkRepositoryFromRoot(t, repoRoot, repoPath)
}

func walkRepositoryFromRoot(t *testing.T, root *os.Root, label string) {
	t.Helper()

	repo, err := repository.Open(root)
	if err != nil {
		t.Fatalf("repository.Open(root for %q): %v", label, err)
	}

	defer func() { _ = repo.Close() }()

	head, err := repo.Refs().ResolveToDetached("HEAD")
	if err != nil {
		t.Fatalf("ResolveRefFully(HEAD): %v", err)
	}

	objectsRead, err := traverseReachableIter(repo, head.ID)
	if err != nil {
		t.Fatalf("traverseReachableIter(%s): %v", head.ID, err)
	}

	if objectsRead <= 0 {
		t.Fatalf("no objects were enumerated from HEAD (%s)", fmt.Sprintf("%q", label))
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

		stored, err := repo.Resolver().ExactObject(id)
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
