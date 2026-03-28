package reachability_test

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"testing"

	giterrors "codeberg.org/lindenii/furgit/errors"
	"codeberg.org/lindenii/furgit/internal/testgit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/store/memory"
	"codeberg.org/lindenii/furgit/object/tree"
	objecttype "codeberg.org/lindenii/furgit/object/type"
	"codeberg.org/lindenii/furgit/reachability"
)

type memStore struct {
	*memory.Store

	readBytesByObjectID map[objectid.ObjectID]int
}

// newCountingMemStore builds one in-memory store that records content-read
// counts by object ID.
func newCountingMemStore(algo objectid.Algorithm) *memStore {
	return &memStore{
		Store:               memory.New(algo),
		readBytesByObjectID: make(map[objectid.ObjectID]int),
	}
}

func (store *memStore) ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error) {
	store.readBytesByObjectID[id]++

	return store.Store.ReadBytesContent(id)
}

func commitBody(tree objectid.ObjectID, parents ...objectid.ObjectID) []byte {
	buf := fmt.Appendf(nil, "tree %s\n", tree.String())
	for _, parent := range parents {
		buf = append(buf, fmt.Appendf(nil, "parent %s\n", parent.String())...)
	}

	buf = append(buf, []byte("\nmsg\n")...)

	return buf
}

func tagBody(target objectid.ObjectID, targetType objecttype.Type) []byte {
	targetName, ok := objecttype.Name(targetType)
	if !ok {
		panic("invalid tag target type")
	}

	return fmt.Appendf(nil, "object %s\ntype %s\ntag t\n\nmsg\n", target.String(), targetName)
}

func collectSeq(seq func(func(objectid.ObjectID) bool)) []objectid.ObjectID {
	var out []objectid.ObjectID

	seq(func(id objectid.ObjectID) bool {
		out = append(out, id)

		return true
	})

	return out
}

func toSet(ids []objectid.ObjectID) map[objectid.ObjectID]struct{} {
	set := make(map[objectid.ObjectID]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}

	return set
}

func TestWalkDomainCommitsIncludesTagNodes(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := newCountingMemStore(algo)
		blob := store.AddObject(objecttype.TypeBlob, []byte("blob\n"))
		tree := store.AddObject(objecttype.TypeTree, mustSerializeTree(t, &tree.Tree{Entries: []tree.TreeEntry{{
			Mode: tree.FileModeRegular,
			Name: []byte("f"),
			ID:   blob,
		}}}))
		commit1 := store.AddObject(objecttype.TypeCommit, commitBody(tree))
		commit2 := store.AddObject(objecttype.TypeCommit, commitBody(tree, commit1))
		tag1 := store.AddObject(objecttype.TypeTag, tagBody(commit2, objecttype.TypeCommit))
		tag2 := store.AddObject(objecttype.TypeTag, tagBody(tag1, objecttype.TypeTag))

		r := reachability.New(store, nil)
		walk := r.Walk(reachability.DomainCommits, nil, map[objectid.ObjectID]struct{}{tag2: {}})

		got := collectSeq(walk.Seq())

		err := walk.Err()
		if err != nil {
			t.Fatalf("walk.Err(): %v", err)
		}

		gotSet := toSet(got)

		wantSet := map[objectid.ObjectID]struct{}{tag2: {}, tag1: {}, commit2: {}, commit1: {}}
		if !maps.Equal(gotSet, wantSet) {
			t.Fatalf("walk output mismatch: got %v, want %v", slices.Collect(maps.Keys(gotSet)), slices.Collect(maps.Keys(wantSet)))
		}
	})
}

func TestWalkExcludesHavesCompletely(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := newCountingMemStore(algo)
		blob := store.AddObject(objecttype.TypeBlob, []byte("blob\n"))
		tree := store.AddObject(objecttype.TypeTree, mustSerializeTree(t, &tree.Tree{Entries: []tree.TreeEntry{{
			Mode: tree.FileModeRegular,
			Name: []byte("f"),
			ID:   blob,
		}}}))
		commit := store.AddObject(objecttype.TypeCommit, commitBody(tree))

		r := reachability.New(store, nil)
		walk := r.Walk(reachability.DomainCommits, map[objectid.ObjectID]struct{}{commit: {}}, map[objectid.ObjectID]struct{}{commit: {}})

		got := collectSeq(walk.Seq())

		err := walk.Err()
		if err != nil {
			t.Fatalf("walk.Err(): %v", err)
		}

		if len(got) != 0 {
			t.Fatalf("expected empty output, got %v", got)
		}
	})
}

func TestWalkDomainCommitsRejectsNonCommitRootAfterPeel(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := newCountingMemStore(algo)
		blob := store.AddObject(objecttype.TypeBlob, []byte("blob\n"))
		tree := store.AddObject(objecttype.TypeTree, mustSerializeTree(t, &tree.Tree{Entries: []tree.TreeEntry{{
			Mode: tree.FileModeRegular,
			Name: []byte("f"),
			ID:   blob,
		}}}))
		tag := store.AddObject(objecttype.TypeTag, tagBody(tree, objecttype.TypeTree))

		r := reachability.New(store, nil)
		walk := r.Walk(reachability.DomainCommits, nil, map[objectid.ObjectID]struct{}{tag: {}})
		_ = collectSeq(walk.Seq())

		err := walk.Err()
		if err == nil {
			t.Fatal("expected error")
		}

		typeErr, ok := errors.AsType[*giterrors.ObjectTypeError](err)
		if !ok {
			t.Fatalf("expected ObjectTypeError, got %T (%v)", err, err)
		}

		if typeErr.Got != objecttype.TypeTree || typeErr.Want != objecttype.TypeCommit {
			t.Fatalf("unexpected type error: %+v", typeErr)
		}
	})
}

func TestWalkDomainCommitsHaveTagStopsTraversal(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := newCountingMemStore(algo)
		blob := store.AddObject(objecttype.TypeBlob, []byte("blob\n"))
		tree := store.AddObject(objecttype.TypeTree, mustSerializeTree(t, &tree.Tree{Entries: []tree.TreeEntry{{
			Mode: tree.FileModeRegular,
			Name: []byte("f"),
			ID:   blob,
		}}}))
		commit1 := store.AddObject(objecttype.TypeCommit, commitBody(tree))
		commit2 := store.AddObject(objecttype.TypeCommit, commitBody(tree, commit1))
		tag1 := store.AddObject(objecttype.TypeTag, tagBody(commit2, objecttype.TypeCommit))
		tag2 := store.AddObject(objecttype.TypeTag, tagBody(tag1, objecttype.TypeTag))

		r := reachability.New(store, nil)
		walk := r.Walk(
			reachability.DomainCommits,
			map[objectid.ObjectID]struct{}{tag1: {}},
			map[objectid.ObjectID]struct{}{tag2: {}},
		)

		got := collectSeq(walk.Seq())

		err := walk.Err()
		if err != nil {
			t.Fatalf("walk.Err(): %v", err)
		}

		gotSet := toSet(got)

		wantSet := map[objectid.ObjectID]struct{}{tag2: {}}
		if !maps.Equal(gotSet, wantSet) {
			t.Fatalf("walk output mismatch: got %v, want %v", slices.Collect(maps.Keys(gotSet)), slices.Collect(maps.Keys(wantSet)))
		}
	})
}

func TestWalkDomainObjectsRecursesTreesAndSkipsBlobContentReads(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := newCountingMemStore(algo)

		blob1 := store.AddObject(objecttype.TypeBlob, []byte("b1\n"))
		blob2 := store.AddObject(objecttype.TypeBlob, []byte("b2\n"))
		gitlinkTarget := store.Algorithm().Sum([]byte("external-submodule"))

		subtree := store.AddObject(objecttype.TypeTree, mustSerializeTree(t, &tree.Tree{Entries: []tree.TreeEntry{{
			Mode: tree.FileModeRegular,
			Name: []byte("nested"),
			ID:   blob2,
		}}}))
		rootTree := store.AddObject(objecttype.TypeTree, mustSerializeTree(t, &tree.Tree{Entries: []tree.TreeEntry{
			{Mode: tree.FileModeRegular, Name: []byte("a"), ID: blob1},
			{Mode: tree.FileModeDir, Name: []byte("dir"), ID: subtree},
			{Mode: tree.FileModeGitlink, Name: []byte("submodule"), ID: gitlinkTarget},
		}}))
		commit := store.AddObject(objecttype.TypeCommit, commitBody(rootTree))

		r := reachability.New(store, nil)
		walk := r.Walk(reachability.DomainObjects, nil, map[objectid.ObjectID]struct{}{commit: {}})

		got := collectSeq(walk.Seq())

		err := walk.Err()
		if err != nil {
			t.Fatalf("walk.Err(): %v", err)
		}

		gotSet := toSet(got)

		wantSet := map[objectid.ObjectID]struct{}{commit: {}, rootTree: {}, subtree: {}, blob1: {}, blob2: {}}
		if !maps.Equal(gotSet, wantSet) {
			t.Fatalf("walk output mismatch: got %v, want %v", slices.Collect(maps.Keys(gotSet)), slices.Collect(maps.Keys(wantSet)))
		}

		if store.readBytesByObjectID[blob1] != 0 || store.readBytesByObjectID[blob2] != 0 {
			t.Fatalf("blob contents should not be read; counts: blob1=%d blob2=%d", store.readBytesByObjectID[blob1], store.readBytesByObjectID[blob2])
		}
	})
}

func TestCheckConnectedReturnsConcreteMissingObject(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := newCountingMemStore(algo)
		blob := store.AddObject(objecttype.TypeBlob, []byte("blob\n"))
		tree := store.AddObject(objecttype.TypeTree, mustSerializeTree(t, &tree.Tree{Entries: []tree.TreeEntry{{
			Mode: tree.FileModeRegular,
			Name: []byte("f"),
			ID:   blob,
		}}}))
		missingParent := store.Algorithm().Sum([]byte("missing-parent"))
		commit := store.AddObject(objecttype.TypeCommit, commitBody(tree, missingParent))

		r := reachability.New(store, nil)

		err := r.CheckConnected(reachability.DomainCommits, nil, map[objectid.ObjectID]struct{}{commit: {}})
		if err == nil {
			t.Fatal("expected error")
		}

		missing, ok := errors.AsType[*giterrors.ObjectMissingError](err)
		if !ok {
			t.Fatalf("expected ObjectMissingError, got %T (%v)", err, err)
		}

		if missing.OID != missingParent {
			t.Fatalf("unexpected missing oid: got %s want %s", missing.OID, missingParent)
		}
	})
}

func TestWalkInvalidDomainReturnsPlainError(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		r := reachability.New(newCountingMemStore(algo), nil)
		walk := r.Walk(reachability.Domain(99), nil, nil)

		_ = collectSeq(walk.Seq())

		err := walk.Err()
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func mustSerializeTree(tb testing.TB, tree *tree.Tree) []byte {
	tb.Helper()

	body, err := tree.SerializeWithoutHeader()
	if err != nil {
		tb.Fatalf("SerializeWithoutHeader: %v", err)
	}

	return body
}
