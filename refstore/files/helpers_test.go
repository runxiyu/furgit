package files_test

import (
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/refstore/files"
)

const testPackedRefsTimeout = time.Second

func openFilesStore(t *testing.T, testRepo *testgit.TestRepo, algo objectid.Algorithm) *files.Store {
	t.Helper()

	root := testRepo.OpenGitRoot(t)

	store, err := files.New(root, algo, testPackedRefsTimeout)
	if err != nil {
		t.Fatalf("files.New: %v", err)
	}

	return store
}

func openFilesStoreAt(t *testing.T, root *os.Root, algo objectid.Algorithm) *files.Store {
	t.Helper()

	store, err := files.New(root, algo, testPackedRefsTimeout)
	if err != nil {
		t.Fatalf("files.New: %v", err)
	}

	return store
}

func openGitRootUnder(t *testing.T, repoRoot *os.Root, worktreeName string) *os.Root {
	t.Helper()

	worktreeRoot, err := repoRoot.OpenRoot(worktreeName)
	if err != nil {
		t.Fatalf("OpenRoot(%q): %v", worktreeName, err)
	}

	t.Cleanup(func() {
		_ = worktreeRoot.Close()
	})

	info, err := worktreeRoot.Stat(".git")
	if err != nil {
		t.Fatalf("stat %q: %v", worktreeName+"/.git", err)
	}

	if info.IsDir() {
		gitRoot, err := worktreeRoot.OpenRoot(".git")
		if err != nil {
			t.Fatalf("OpenRoot(.git): %v", err)
		}

		t.Cleanup(func() {
			_ = gitRoot.Close()
		})

		return gitRoot
	}

	content, err := worktreeRoot.ReadFile(".git")
	if err != nil {
		t.Fatalf("read %q: %v", worktreeName+"/.git", err)
	}

	gitDir := strings.TrimSpace(strings.TrimPrefix(string(content), "gitdir:"))
	if gitDir == "" {
		t.Fatalf("%q does not contain a gitdir path", worktreeName+"/.git")
	}

	if strings.HasPrefix(gitDir, "/") {
		gitRoot, err := os.OpenRoot(gitDir)
		if err != nil {
			t.Fatalf("os.OpenRoot(%q): %v", gitDir, err)
		}

		t.Cleanup(func() {
			_ = gitRoot.Close()
		})

		return gitRoot
	}

	gitRoot, err := worktreeRoot.OpenRoot(gitDir)
	if err != nil {
		t.Fatalf("os.OpenRoot(%q): %v", gitDir, err)
	}

	t.Cleanup(func() {
		_ = gitRoot.Close()
	})

	return gitRoot
}

func assertListMatchesGitForEachRef(t *testing.T, gitOut string, store *files.Store) {
	t.Helper()

	listed, err := store.List("")
	if err != nil {
		t.Fatalf("List(\"\"): %v", err)
	}

	gotNames := make([]string, 0, len(listed))
	for _, got := range listed {
		if got.Name() == "HEAD" {
			continue
		}

		gotNames = append(gotNames, got.Name())
	}

	slices.Sort(gotNames)

	wantLines := strings.Split(strings.TrimSpace(gitOut), "\n")
	wantNames := make([]string, 0, len(wantLines))

	for _, line := range wantLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		wantNames = append(wantNames, line)
	}

	slices.Sort(wantNames)

	if !slices.Equal(gotNames, wantNames) {
		t.Fatalf("List names = %v, want %v", gotNames, wantNames)
	}
}

func forEachRefLines(output string) []string {
	if strings.TrimSpace(output) == "" {
		return nil
	}

	return strings.Split(strings.TrimSpace(output), "\n")
}
