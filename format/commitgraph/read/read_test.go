package read_test

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/format/commitgraph/bloom"
	"codeberg.org/lindenii/furgit/format/commitgraph/read"
	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/internal/testgit"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

func fixtureRepoPath(t *testing.T, algo objectid.Algorithm, name string) string {
	t.Helper()

	return filepath.Join("testdata", "fixtures", algo.String(), name, "repo.git")
}

func fixtureRepo(t *testing.T, algo objectid.Algorithm, name string) *testgit.TestRepo {
	t.Helper()

	return testgit.NewRepoFromFixture(t, algo, fixtureRepoPath(t, algo, name))
}

func TestReadSingleMatchesGit(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := fixtureRepo(t, algo, "single_changed")

		reader := openReader(t, testRepo, read.OpenSingle)

		defer func() { _ = reader.Close() }()

		allIDs := testRepo.RevList(t, "--all")
		if len(allIDs) == 0 {
			t.Fatal("git rev-list --all returned no commits")
		}

		wantCommitCount, err := intconv.IntToUint32(len(allIDs))
		if err != nil {
			t.Fatalf("len(allIDs) convert: %v", err)
		}

		if got := reader.NumCommits(); got != wantCommitCount {
			t.Fatalf("NumCommits() = %d, want %d", got, len(allIDs))
		}

		if !reader.HasBloom() {
			t.Fatal("HasBloom() = false, want true")
		}

		bloomVersion := reader.BloomVersion()
		if bloomVersion == 0 {
			t.Fatal("BloomVersion() = 0, want non-zero when HasBloom() is true")
		}

		for _, id := range allIDs {
			pos, err := reader.Lookup(id)
			if err != nil {
				t.Fatalf("Lookup(%s): %v", id, err)
			}

			gotID, err := reader.OIDAt(pos)
			if err != nil {
				t.Fatalf("OIDAt(%+v): %v", pos, err)
			}

			if gotID != id {
				t.Fatalf("OIDAt(Lookup(%s)) = %s, want %s", id, gotID, id)
			}
		}

		step := max(len(allIDs)/24, 1)

		for i, id := range allIDs {
			if i%step != 0 && i != len(allIDs)-1 {
				continue
			}

			verifyCommitAgainstGit(t, testRepo, reader, id)
		}
	})
}

func TestReadChainMatchesGit(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := fixtureRepo(t, algo, "chain_changed")

		reader := openReader(t, testRepo, read.OpenChain)

		defer func() { _ = reader.Close() }()

		layers := reader.Layers()
		if len(layers) < 2 {
			t.Fatalf("Layers len = %d, want >= 2", len(layers))
		}

		allIDs := testRepo.RevList(t, "--all")

		wantCommitCount, err := intconv.IntToUint32(len(allIDs))
		if err != nil {
			t.Fatalf("len(allIDs) convert: %v", err)
		}

		if got := reader.NumCommits(); got != wantCommitCount {
			t.Fatalf("NumCommits() = %d, want %d", got, len(allIDs))
		}

		step := max(len(allIDs)/20, 1)

		for i, id := range allIDs {
			pos, err := reader.Lookup(id)
			if err != nil {
				t.Fatalf("Lookup(%s): %v", id, err)
			}

			if i%step != 0 && i != len(allIDs)-1 {
				continue
			}

			gotID, err := reader.OIDAt(pos)
			if err != nil {
				t.Fatalf("OIDAt(%+v): %v", pos, err)
			}

			if gotID != id {
				t.Fatalf("OIDAt(Lookup(%s)) = %s, want %s", id, gotID, id)
			}
		}
	})
}

func TestBloomUnavailableWithoutChangedPaths(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := fixtureRepo(t, algo, "single_nochanged")

		reader := openReader(t, testRepo, read.OpenSingle)

		defer func() { _ = reader.Close() }()

		head := testRepo.RevParse(t, "HEAD")

		pos, err := reader.Lookup(head)
		if err != nil {
			t.Fatalf("Lookup(%s): %v", head, err)
		}

		_, err = reader.BloomFilterAt(pos)
		if err == nil {
			t.Fatal("BloomFilterAt() error = nil, want BloomUnavailableError")
		}

		unavailable, ok := errors.AsType[*read.BloomUnavailableError](err)
		if !ok {
			t.Fatalf("BloomFilterAt() error type = %T, want *BloomUnavailableError", err)
		}

		if unavailable.Pos != pos {
			t.Fatalf("BloomUnavailableError.Pos = %+v, want %+v", unavailable.Pos, pos)
		}
	})
}

func openReader(tb testing.TB, testRepo *testgit.TestRepo, mode read.OpenMode) *read.Reader {
	tb.Helper()

	root := testRepo.OpenObjectsRoot(tb)

	reader, err := read.Open(root, testRepo.Algorithm(), mode)
	if err != nil {
		tb.Fatalf("read.Open(objects): %v", err)
	}

	return reader
}

func verifyCommitAgainstGit(tb testing.TB, testRepo *testgit.TestRepo, reader *read.Reader, id objectid.ObjectID) {
	tb.Helper()

	pos, err := reader.Lookup(id)
	if err != nil {
		tb.Fatalf("Lookup(%s): %v", id, err)
	}

	commit, err := reader.CommitAt(pos)
	if err != nil {
		tb.Fatalf("CommitAt(%+v): %v", pos, err)
	}

	if commit.OID != id {
		tb.Fatalf("CommitAt(%+v).OID = %s, want %s", pos, commit.OID, id)
	}

	treeHex := testRepo.Run(tb, "show", "-s", "--format=%T", id.String())

	wantTree, err := objectid.ParseHex(testRepo.Algorithm(), treeHex)
	if err != nil {
		tb.Fatalf("parse tree id %q: %v", treeHex, err)
	}

	if commit.TreeOID != wantTree {
		tb.Fatalf("CommitAt(%+v).TreeOID = %s, want %s", pos, commit.TreeOID, wantTree)
	}

	wantParents := parseOIDLine(tb, testRepo.Algorithm(), testRepo.Run(tb, "show", "-s", "--format=%P", id.String()))

	gotParents := commitParents(tb, reader, commit)
	if len(gotParents) != len(wantParents) {
		tb.Fatalf("parent count for %s = %d, want %d", id, len(gotParents), len(wantParents))
	}

	for i := range gotParents {
		if gotParents[i] != wantParents[i] {
			tb.Fatalf("parent %d for %s = %s, want %s", i, id, gotParents[i], wantParents[i])
		}
	}

	commitTimeRaw := testRepo.Run(tb, "show", "-s", "--format=%ct", id.String())

	wantCommitTime, err := strconv.ParseInt(strings.TrimSpace(commitTimeRaw), 10, 64)
	if err != nil {
		tb.Fatalf("parse commit time %q: %v", commitTimeRaw, err)
	}

	if commit.CommitTimeUnix != wantCommitTime {
		tb.Fatalf("CommitAt(%+v).CommitTimeUnix = %d, want %d", pos, commit.CommitTimeUnix, wantCommitTime)
	}

	filter, err := reader.BloomFilterAt(pos)
	if err != nil {
		tb.Fatalf("BloomFilterAt(%+v): %v", pos, err)
	}

	if filter.HashVersion != uint32(reader.BloomVersion()) {
		tb.Fatalf("filter.HashVersion = %d, want %d", filter.HashVersion, reader.BloomVersion())
	}

	assertChangedPathsBloomPositive(tb, testRepo, filter, id)
}

func commitParents(tb testing.TB, reader *read.Reader, commit read.Commit) []objectid.ObjectID {
	tb.Helper()

	out := make([]objectid.ObjectID, 0, 2+len(commit.ExtraParents))

	if commit.Parent1.Valid {
		id, err := reader.OIDAt(commit.Parent1.Pos)
		if err != nil {
			tb.Fatalf("OIDAt(parent1 %+v): %v", commit.Parent1.Pos, err)
		}

		out = append(out, id)
	}

	if commit.Parent2.Valid {
		id, err := reader.OIDAt(commit.Parent2.Pos)
		if err != nil {
			tb.Fatalf("OIDAt(parent2 %+v): %v", commit.Parent2.Pos, err)
		}

		out = append(out, id)
	}

	for _, parentPos := range commit.ExtraParents {
		id, err := reader.OIDAt(parentPos)
		if err != nil {
			tb.Fatalf("OIDAt(extra parent %+v): %v", parentPos, err)
		}

		out = append(out, id)
	}

	return out
}

func assertChangedPathsBloomPositive(tb testing.TB, testRepo *testgit.TestRepo, filter bloom.Filter, commitID objectid.ObjectID) {
	tb.Helper()

	changedPaths := testRepo.Run(tb, "diff-tree", "--no-commit-id", "--name-only", "-r", "--root", commitID.String())
	for line := range strings.SplitSeq(strings.TrimSpace(changedPaths), "\n") {
		path := strings.TrimSpace(line)
		if path == "" {
			continue
		}

		mightContain, err := filter.MightContain([]byte(path))
		if err != nil {
			tb.Fatalf("MightContain(%q): %v", path, err)
		}

		if !mightContain {
			tb.Fatalf("Bloom filter false negative for commit %s path %q", commitID, path)
		}
	}
}

func parseOIDLine(tb testing.TB, algo objectid.Algorithm, line string) []objectid.ObjectID {
	tb.Helper()

	toks := strings.Fields(line)

	out := make([]objectid.ObjectID, 0, len(toks))
	for _, tok := range toks {
		id, err := objectid.ParseHex(algo, tok)
		if err != nil {
			tb.Fatalf("parse object id %q: %v", tok, err)
		}

		out = append(out, id)
	}

	return out
}
