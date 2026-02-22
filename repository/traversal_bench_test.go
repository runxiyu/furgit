package repository_test

import (
	"os"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/repository"
)

const benchRepoPathEnv = "FURGIT_BENCH_REPO"

// BenchmarkTraverseHeadTree measures iterative traversal of HEAD's root tree.
//
// Set FURGIT_BENCH_REPO to a repository path (typically .git or a bare repo)
// before running this benchmark.
func BenchmarkTraverseHeadTree(b *testing.B) {
	repoPath := strings.TrimSpace(os.Getenv(benchRepoPathEnv))
	if repoPath == "" {
		b.Fatalf("missing %s", benchRepoPathEnv)
	}

	root, err := os.OpenRoot(repoPath)
	if err != nil {
		b.Fatalf("os.OpenRoot(%q): %v", repoPath, err)
	}
	b.Cleanup(func() {
		_ = root.Close()
	})

	repo, err := repository.Open(root)
	if err != nil {
		b.Fatalf("repository.Open(root for %q): %v", repoPath, err)
	}
	b.Cleanup(func() {
		_ = repo.Close()
	})

	head, err := repo.ResolveRefFully("HEAD")
	if err != nil {
		b.Fatalf("ResolveRefFully(HEAD): %v", err)
	}
	stored, err := repo.ReadStored(head.ID)
	if err != nil {
		b.Fatalf("ReadStored(%s): %v", head.ID, err)
	}
	commit, ok := stored.Object().(*object.Commit)
	if !ok {
		b.Fatalf("HEAD object type %T, want *object.Commit", stored.Object())
	}

	b.ReportAllocs()
	b.ResetTimer()

	var lastCount int
	for b.Loop() {
		lastCount, err = traverseTreeIter(repo, commit.Tree)
		if err != nil {
			b.Fatalf("traverseTreeIter: %v", err)
		}
	}

	b.StopTimer()
	if lastCount <= 0 {
		b.Fatalf("traverseTreeIter count = %d, want > 0", lastCount)
	}
}
