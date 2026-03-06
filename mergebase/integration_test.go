package mergebase_test

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/mergebase"
	"codeberg.org/lindenii/furgit/objectid"
)

func TestQueryMatchesGitMergeBaseAll(t *testing.T) {
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

		tag := testRepo.TagAnnotated(t, "right-tag", right, "right-tag")

		store := testRepo.OpenObjectStore(t)

		query := mergebase.Query(store, nil, left, tag)
		got := oidSetFromSeq(query.Seq())

		err := query.Err()
		if err != nil {
			t.Fatalf("query.Err(): %v", err)
		}

		want := gitMergeBaseAllSet(t, testRepo, left, tag)
		if !maps.Equal(got, want) {
			t.Fatalf("Query(left, tag) mismatch:\n got=%v\nwant=%v", sortedOIDStrings(got), sortedOIDStrings(want))
		}
	})
}

func TestQueryCrissCrossMatchesGitMergeBaseAll(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		_, tree1 := testRepo.MakeSingleFileTree(t, "root.txt", []byte("root\n"))
		root := testRepo.CommitTree(t, tree1, "root")

		_, tree2 := testRepo.MakeSingleFileTree(t, "base1.txt", []byte("base1\n"))
		base1 := testRepo.CommitTree(t, tree2, "base1", root)

		_, tree3 := testRepo.MakeSingleFileTree(t, "base2.txt", []byte("base2\n"))
		base2 := testRepo.CommitTree(t, tree3, "base2", root)

		_, tree4 := testRepo.MakeSingleFileTree(t, "left.txt", []byte("left\n"))
		left := testRepo.CommitTree(t, tree4, "left", base1, base2)

		_, tree5 := testRepo.MakeSingleFileTree(t, "right.txt", []byte("right\n"))
		right := testRepo.CommitTree(t, tree5, "right", base2, base1)

		store := testRepo.OpenObjectStore(t)

		query := mergebase.Query(store, nil, left, right)
		got := oidSetFromSeq(query.Seq())

		err := query.Err()
		if err != nil {
			t.Fatalf("query.Err(): %v", err)
		}

		want := gitMergeBaseAllSet(t, testRepo, left, right)
		if !maps.Equal(got, want) {
			t.Fatalf("Query(left, right) mismatch:\n got=%v\nwant=%v", sortedOIDStrings(got), sortedOIDStrings(want))
		}

		first, ok, err := mergebase.Base(store, nil, left, right)
		if err != nil {
			t.Fatalf("Base(left, right): %v", err)
		}

		if !ok {
			t.Fatal("Base(left, right) unexpectedly reported no base")
		}

		if !containsID(want, first) {
			t.Fatalf("Base(left, right)=%s, want one of %v", first, slices.Collect(maps.Keys(want)))
		}
	})
}

func TestQueryMatchesGitMergeBaseAllWithCommitGraph(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		_, tree1 := testRepo.MakeSingleFileTree(t, "root.txt", []byte("root\n"))
		root := testRepo.CommitTree(t, tree1, "root")

		_, tree2 := testRepo.MakeSingleFileTree(t, "base1.txt", []byte("base1\n"))
		base1 := testRepo.CommitTree(t, tree2, "base1", root)

		_, tree3 := testRepo.MakeSingleFileTree(t, "base2.txt", []byte("base2\n"))
		base2 := testRepo.CommitTree(t, tree3, "base2", root)

		_, tree4 := testRepo.MakeSingleFileTree(t, "left.txt", []byte("left\n"))
		left := testRepo.CommitTree(t, tree4, "left", base1, base2)

		_, tree5 := testRepo.MakeSingleFileTree(t, "right.txt", []byte("right\n"))
		right := testRepo.CommitTree(t, tree5, "right", base2, base1)

		testRepo.UpdateRef(t, "refs/heads/main", right)
		testRepo.SymbolicRef(t, "HEAD", "refs/heads/main")
		testRepo.CommitGraphWrite(t, "--reachable")

		store := testRepo.OpenObjectStore(t)
		graph := testRepo.OpenCommitGraph(t)

		query := mergebase.Query(store, graph, left, right)
		got := oidSetFromSeq(query.Seq())

		err := query.Err()
		if err != nil {
			t.Fatalf("query.Err(): %v", err)
		}

		want := gitMergeBaseAllSet(t, testRepo, left, right)
		if !maps.Equal(got, want) {
			t.Fatalf("Query(left, right) with commit-graph mismatch:\n got=%v\nwant=%v", sortedOIDStrings(got), sortedOIDStrings(want))
		}

		first, ok, err := mergebase.Base(store, graph, left, right)
		if err != nil {
			t.Fatalf("Base(left, right): %v", err)
		}

		if !ok {
			t.Fatal("Base(left, right) unexpectedly reported no base")
		}

		if !containsID(want, first) {
			t.Fatalf("Base(left, right)=%s, want one of %v", first, slices.Collect(maps.Keys(want)))
		}
	})
}

func TestBaseMatchesGitMergeBaseWithoutAll(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		_, tree1 := testRepo.MakeSingleFileTree(t, "root.txt", []byte("root\n"))
		root := testRepo.CommitTree(t, tree1, "root")

		_, tree2 := testRepo.MakeSingleFileTree(t, "base1.txt", []byte("base1\n"))
		base1 := testRepo.CommitTreeWithEnv(t, []string{
			"GIT_AUTHOR_DATE=1234567890 +0000",
			"GIT_COMMITTER_DATE=1234567890 +0000",
		}, tree2, "base1", root)

		_, tree3 := testRepo.MakeSingleFileTree(t, "base2.txt", []byte("base2\n"))
		base2 := testRepo.CommitTreeWithEnv(t, []string{
			"GIT_AUTHOR_DATE=1234567990 +0000",
			"GIT_COMMITTER_DATE=1234567990 +0000",
		}, tree3, "base2", root)

		_, tree4 := testRepo.MakeSingleFileTree(t, "left.txt", []byte("left\n"))
		left := testRepo.CommitTree(t, tree4, "left", base1, base2)

		_, tree5 := testRepo.MakeSingleFileTree(t, "right.txt", []byte("right\n"))
		right := testRepo.CommitTree(t, tree5, "right", base2, base1)

		store := testRepo.OpenObjectStore(t)

		got, ok, err := mergebase.Base(store, nil, left, right)
		if err != nil {
			t.Fatalf("Base(left, right): %v", err)
		}

		if !ok {
			t.Fatal("Base(left, right) unexpectedly reported no base")
		}

		want := gitMergeBaseOne(t, testRepo, left, right)
		if got != want {
			t.Fatalf("Base(left, right)=%s, want %s", got, want)
		}

		testRepo.UpdateRef(t, "refs/heads/main", right)
		testRepo.SymbolicRef(t, "HEAD", "refs/heads/main")
		testRepo.CommitGraphWrite(t, "--reachable")

		graph := testRepo.OpenCommitGraph(t)

		got, ok, err = mergebase.Base(store, graph, left, right)
		if err != nil {
			t.Fatalf("Base(left, right) with commit-graph: %v", err)
		}

		if !ok {
			t.Fatal("Base(left, right) with commit-graph unexpectedly reported no base")
		}

		if got != want {
			t.Fatalf("Base(left, right) with commit-graph=%s, want %s", got, want)
		}
	})
}

// oidSetFromSeq collects one object ID sequence into a set.
func oidSetFromSeq(seq func(func(objectid.ObjectID) bool)) map[objectid.ObjectID]struct{} {
	out := make(map[objectid.ObjectID]struct{})

	seq(func(id objectid.ObjectID) bool {
		out[id] = struct{}{}

		return true
	})

	return out
}

// gitMergeBaseAllSet returns Git's merge-base --all output as a set.
func gitMergeBaseAllSet(
	t *testing.T,
	testRepo *testgit.TestRepo,
	left objectid.ObjectID,
	right objectid.ObjectID,
) map[objectid.ObjectID]struct{} {
	t.Helper()

	out := testRepo.Run(t, "merge-base", "--all", left.String(), right.String())
	set := make(map[objectid.ObjectID]struct{})

	for line := range strings.SplitSeq(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		id, err := objectid.ParseHex(testRepo.Algorithm(), line)
		if err != nil {
			t.Fatalf("parse merge-base oid %q: %v", line, err)
		}

		set[id] = struct{}{}
	}

	return set
}

// gitMergeBaseOne returns Git's merge-base output without --all.
func gitMergeBaseOne(
	t *testing.T,
	testRepo *testgit.TestRepo,
	left objectid.ObjectID,
	right objectid.ObjectID,
) objectid.ObjectID {
	t.Helper()

	out := strings.TrimSpace(testRepo.Run(t, "merge-base", left.String(), right.String()))
	if out == "" {
		t.Fatal("git merge-base returned no output")
	}

	id, err := objectid.ParseHex(testRepo.Algorithm(), out)
	if err != nil {
		t.Fatalf("parse merge-base oid %q: %v", out, err)
	}

	return id
}

func sortedOIDStrings(set map[objectid.ObjectID]struct{}) []string {
	out := make([]string, 0, len(set))
	for id := range set {
		out = append(out, id.String())
	}

	slices.Sort(out)

	return out
}
