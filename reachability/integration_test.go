package reachability_test

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/reachability"
	"codeberg.org/lindenii/furgit/repository"
)

func TestWalkCommitsMatchesGitRevList(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		_, tree1 := testRepo.MakeSingleFileTree(t, "base.txt", []byte("base\n"))
		base := testRepo.CommitTree(t, tree1, "base")

		_, tree2 := testRepo.MakeSingleFileTree(t, "left.txt", []byte("left\n"))
		left := testRepo.CommitTree(t, tree2, "left", base)

		_, tree3 := testRepo.MakeSingleFileTree(t, "right.txt", []byte("right\n"))
		right := testRepo.CommitTree(t, tree3, "right", base)

		_, tree4 := testRepo.MakeSingleFileTree(t, "merge.txt", []byte("merge\n"))
		merge := testRepo.CommitTree(t, tree4, "merge", left, right)

		tag1 := testRepo.TagAnnotated(t, "v1", merge, "v1")
		tag2 := testRepo.TagAnnotated(t, "v2", tag1, "v2")

		r := openReachabilityFromTestRepo(t, testRepo)
		walk := r.Walk(
			reachability.DomainCommits,
			nil,
			map[objectid.ObjectID]struct{}{merge: {}},
		)

		got := oidSetFromSeq(walk.Seq())

		err := walk.Err()
		if err != nil {
			t.Fatalf("walk.Err(): %v", err)
		}

		want := gitRevListSet(t, testRepo, false, []objectid.ObjectID{merge}, nil)
		if !maps.Equal(got, want) {
			t.Fatalf("commit walk mismatch:\n got=%v\nwant=%v", sortedOIDStrings(got), sortedOIDStrings(want))
		}

		peelWalk := r.Walk(
			reachability.DomainCommits,
			nil,
			map[objectid.ObjectID]struct{}{tag2: {}},
		)

		peelGot := oidSetFromSeq(peelWalk.Seq())

		err = peelWalk.Err()
		if err != nil {
			t.Fatalf("peelWalk.Err(): %v", err)
		}

		wantWithTags := maps.Clone(want)
		wantWithTags[tag1] = struct{}{}

		wantWithTags[tag2] = struct{}{}
		if !maps.Equal(peelGot, wantWithTags) {
			t.Fatalf("tag-root commit walk mismatch:\n got=%v\nwant=%v", sortedOIDStrings(peelGot), sortedOIDStrings(wantWithTags))
		}
	})
}

func TestWalkObjectsMatchesGitRevListObjects(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		aBlob := testRepo.HashObject(t, "blob", []byte("a\n"))
		bBlob := testRepo.HashObject(t, "blob", []byte("b\n"))
		nestedTree := testRepo.Mktree(t, fmt.Sprintf("100644 blob %s\tb.txt\n", bBlob))
		rootTree := testRepo.Mktree(t,
			fmt.Sprintf("100644 blob %s\ta.txt\n040000 tree %s\tdir\n", aBlob, nestedTree),
		)
		base := testRepo.CommitTree(t, rootTree, "base")

		cBlob := testRepo.HashObject(t, "blob", []byte("c\n"))
		tree2 := testRepo.Mktree(t, fmt.Sprintf("100644 blob %s\tc.txt\n", cBlob))
		head := testRepo.CommitTree(t, tree2, "head", base)
		tag := testRepo.TagAnnotated(t, "objtag", head, "objtag")

		r := openReachabilityFromTestRepo(t, testRepo)
		walk := r.Walk(
			reachability.DomainObjects,
			nil,
			map[objectid.ObjectID]struct{}{head: {}},
		)

		got := oidSetFromSeq(walk.Seq())

		err := walk.Err()
		if err != nil {
			t.Fatalf("walk.Err(): %v", err)
		}

		want := gitRevListSet(t, testRepo, true, []objectid.ObjectID{head}, nil)
		if !maps.Equal(got, want) {
			t.Fatalf("object walk mismatch:\n got=%v\nwant=%v", sortedOIDStrings(got), sortedOIDStrings(want))
		}

		peelWalk := r.Walk(
			reachability.DomainObjects,
			nil,
			map[objectid.ObjectID]struct{}{tag: {}},
		)

		peelGot := oidSetFromSeq(peelWalk.Seq())

		err = peelWalk.Err()
		if err != nil {
			t.Fatalf("peelWalk.Err(): %v", err)
		}

		wantFromTag := gitRevListSet(t, testRepo, true, []objectid.ObjectID{tag}, nil)
		if !maps.Equal(peelGot, wantFromTag) {
			t.Fatalf("tag-root object walk mismatch:\n got=%v\nwant=%v", sortedOIDStrings(peelGot), sortedOIDStrings(wantFromTag))
		}

		walkWithHave := r.Walk(
			reachability.DomainObjects,
			map[objectid.ObjectID]struct{}{base: {}},
			map[objectid.ObjectID]struct{}{head: {}},
		)

		withHave := oidSetFromSeq(walkWithHave.Seq())

		err = walkWithHave.Err()
		if err != nil {
			t.Fatalf("walkWithHave.Err(): %v", err)
		}

		_, ok := withHave[base]
		if ok {
			t.Fatalf("walk output unexpectedly contains have commit %s", base)
		}
	})
}

func TestIsAncestorMatchesGitMergeBase(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		_, tree1 := testRepo.MakeSingleFileTree(t, "one.txt", []byte("one\n"))
		c1 := testRepo.CommitTree(t, tree1, "c1")

		_, tree2 := testRepo.MakeSingleFileTree(t, "two.txt", []byte("two\n"))
		c2 := testRepo.CommitTree(t, tree2, "c2", c1)

		_, tree3 := testRepo.MakeSingleFileTree(t, "three.txt", []byte("three\n"))
		c3 := testRepo.CommitTree(t, tree3, "c3", c2)

		tag := testRepo.TagAnnotated(t, "tip", c2, "tip")

		r := openReachabilityFromTestRepo(t, testRepo)

		got, err := r.IsAncestor(c1, tag)
		if err != nil {
			t.Fatalf("IsAncestor(c1, tag): %v", err)
		}

		want := gitMergeBaseIsAncestor(t, testRepo, c1, c2)
		if got != want {
			t.Fatalf("IsAncestor(c1, tag)=%v, want %v", got, want)
		}

		got, err = r.IsAncestor(c3, c2)
		if err != nil {
			t.Fatalf("IsAncestor(c3, c2): %v", err)
		}

		want = gitMergeBaseIsAncestor(t, testRepo, c3, c2)
		if got != want {
			t.Fatalf("IsAncestor(c3, c2)=%v, want %v", got, want)
		}
	})
}

func TestCheckConnectedMissingObject(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		_, treeID, commitID := testRepo.MakeCommit(t, "missing")

		err := os.Remove(looseObjectPath(testRepo.Dir(), treeID))
		if err != nil {
			t.Fatalf("remove tree object: %v", err)
		}

		r := openReachabilityFromTestRepo(t, testRepo)

		err = r.CheckConnected(
			reachability.DomainObjects,
			nil,
			map[objectid.ObjectID]struct{}{commitID: {}},
		)
		if err == nil {
			t.Fatal("expected error")
		}

		var missing *reachability.ErrObjectMissing
		if !errors.As(err, &missing) {
			t.Fatalf("expected ErrObjectMissing, got %T (%v)", err, err)
		}

		if missing.OID != treeID {
			t.Fatalf("missing oid = %s, want %s", missing.OID, treeID)
		}
	})
}

func TestWalkOnPackedOnlyRepo(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		_, tree1 := testRepo.MakeSingleFileTree(t, "one.txt", []byte("one\n"))
		c1 := testRepo.CommitTree(t, tree1, "one")
		_, tree2 := testRepo.MakeSingleFileTree(t, "two.txt", []byte("two\n"))
		c2 := testRepo.CommitTree(t, tree2, "two", c1)
		testRepo.UpdateRef(t, "refs/heads/main", c2)
		testRepo.SymbolicRef(t, "HEAD", "refs/heads/main")

		testRepo.Repack(t, "-ad")
		testRepo.Run(t, "prune-packed")

		assertPackedOnly(t, testRepo.Dir())

		r := openReachabilityFromTestRepo(t, testRepo)
		walk := r.Walk(
			reachability.DomainCommits,
			nil,
			map[objectid.ObjectID]struct{}{c2: {}},
		)

		got := oidSetFromSeq(walk.Seq())

		err := walk.Err()
		if err != nil {
			t.Fatalf("walk.Err(): %v", err)
		}

		_, ok := got[c2]
		if !ok {
			t.Fatalf("walk output missing HEAD commit %s", c2)
		}

		_, ok = got[c1]
		if !ok {
			t.Fatalf("walk output missing parent commit %s", c1)
		}
	})
}

func openReachabilityFromTestRepo(t *testing.T, testRepo *testgit.TestRepo) *reachability.Reachability {
	t.Helper()

	root, err := os.OpenRoot(testRepo.Dir())
	if err != nil {
		t.Fatalf("os.OpenRoot: %v", err)
	}

	t.Cleanup(func() { _ = root.Close() })

	repo, err := repository.Open(root)
	if err != nil {
		t.Fatalf("repository.Open: %v", err)
	}

	t.Cleanup(func() { _ = repo.Close() })

	return reachability.New(repo.Objects())
}

func oidSetFromSeq(seq func(func(objectid.ObjectID) bool)) map[objectid.ObjectID]struct{} {
	out := make(map[objectid.ObjectID]struct{})

	seq(func(id objectid.ObjectID) bool {
		out[id] = struct{}{}

		return true
	})

	return out
}

func gitRevListSet(
	t *testing.T,
	testRepo *testgit.TestRepo,
	includeObjects bool,
	wants []objectid.ObjectID,
	haves []objectid.ObjectID,
) map[objectid.ObjectID]struct{} {
	t.Helper()

	args := []string{"rev-list"}
	if includeObjects {
		args = append(args, "--objects")
	}

	for _, want := range wants {
		args = append(args, want.String())
	}

	if len(haves) > 0 {
		args = append(args, "--not")
		for _, have := range haves {
			args = append(args, have.String())
		}
	}

	out := testRepo.Run(t, args...)
	set := make(map[objectid.ObjectID]struct{})

	for line := range strings.SplitSeq(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		tok := line

		i := strings.IndexByte(tok, ' ')
		if i >= 0 {
			tok = tok[:i]
		}

		id, err := objectid.ParseHex(testRepo.Algorithm(), tok)
		if err != nil {
			t.Fatalf("parse rev-list oid %q: %v", tok, err)
		}

		set[id] = struct{}{}
	}

	return set
}

func gitMergeBaseIsAncestor(t *testing.T, testRepo *testgit.TestRepo, a, b objectid.ObjectID) bool {
	t.Helper()
	// testgit.Run fatals on non-zero status, so we compare merge-base output.
	mb := testRepo.Run(t, "merge-base", a.String(), b.String())

	return mb == a.String()
}

func sortedOIDStrings(set map[objectid.ObjectID]struct{}) []string {
	out := make([]string, 0, len(set))
	for id := range set {
		out = append(out, id.String())
	}

	slices.Sort(out)

	return out
}

func looseObjectPath(repoDir string, id objectid.ObjectID) string {
	hex := id.String()

	return filepath.Join(repoDir, "objects", hex[:2], hex[2:])
}

func assertPackedOnly(t *testing.T, repoDir string) {
	t.Helper()

	objectsDir := filepath.Join(repoDir, "objects")

	entries, err := os.ReadDir(objectsDir)
	if err != nil {
		t.Fatalf("ReadDir(objects): %v", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if name == "pack" || name == "info" {
			continue
		}

		if len(name) == 2 && isHexDirName(name) {
			subEntries, err := os.ReadDir(filepath.Join(objectsDir, name))
			if err != nil {
				t.Fatalf("ReadDir(objects/%s): %v", name, err)
			}

			if len(subEntries) != 0 {
				t.Fatalf("found loose objects in %s", filepath.Join(objectsDir, name))
			}
		}
	}
}

func isHexDirName(name string) bool {
	if len(name) != 2 {
		return false
	}

	for i := range 2 {
		c := name[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}

	return true
}
