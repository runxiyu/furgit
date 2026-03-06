package mergebase_test

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"testing"

	giterrors "codeberg.org/lindenii/furgit/errors"
	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/mergebase"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore/memory"
	"codeberg.org/lindenii/furgit/objecttype"
)

// commitBody serializes one minimal commit body.
func commitBody(tree objectid.ObjectID, parents ...objectid.ObjectID) []byte {
	buf := fmt.Appendf(nil, "tree %s\n", tree.String())
	for _, parent := range parents {
		buf = append(buf, fmt.Appendf(nil, "parent %s\n", parent.String())...)
	}

	buf = append(buf, []byte("\nmsg\n")...)

	return buf
}

// tagBody serializes one minimal annotated tag body.
func tagBody(target objectid.ObjectID, targetType objecttype.Type) []byte {
	targetName, ok := objecttype.Name(targetType)
	if !ok {
		panic("invalid tag target type")
	}

	return fmt.Appendf(nil, "object %s\ntype %s\ntag t\n\nmsg\n", target.String(), targetName)
}

// collectSeq collects one object ID sequence into a slice.
func collectSeq(seq func(func(objectid.ObjectID) bool)) []objectid.ObjectID {
	var out []objectid.ObjectID

	seq(func(id objectid.ObjectID) bool {
		out = append(out, id)

		return true
	})

	return out
}

// toSet converts one slice of object IDs into a set.
func toSet(ids []objectid.ObjectID) map[objectid.ObjectID]struct{} {
	set := make(map[objectid.ObjectID]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}

	return set
}

// containsID reports whether one set contains one object ID.
func containsID(set map[objectid.ObjectID]struct{}, id objectid.ObjectID) bool {
	_, ok := set[id]

	return ok
}

// mustSerializeTree serializes one tree or fails the test.
func mustSerializeTree(tb testing.TB, tree *object.Tree) []byte {
	tb.Helper()

	body, err := tree.SerializeWithoutHeader()
	if err != nil {
		tb.Fatalf("SerializeWithoutHeader: %v", err)
	}

	return body
}

// TestQueryLinearHistory reports one linear-history merge base.
func TestQueryLinearHistory(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := memory.New(algo)
		blob := store.AddObject(objecttype.TypeBlob, []byte("blob\n"))
		tree := store.AddObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("f"),
			ID:   blob,
		}}}))
		base := store.AddObject(objecttype.TypeCommit, commitBody(tree))
		left := store.AddObject(objecttype.TypeCommit, commitBody(tree, base))
		right := store.AddObject(objecttype.TypeCommit, commitBody(tree, left))

		query := mergebase.Query(store, nil, left, right)
		got := collectSeq(query.Seq())

		err := query.Err()
		if err != nil {
			t.Fatalf("query.Err(): %v", err)
		}

		if !slices.Equal(got, []objectid.ObjectID{left}) {
			t.Fatalf("Query(left, right)=%v, want [%s]", got, left)
		}

		first, ok, err := mergebase.Base(store, nil, left, right)
		if err != nil {
			t.Fatalf("Base(left, right): %v", err)
		}

		if !ok {
			t.Fatal("Base(left, right) unexpectedly reported no base")
		}

		if first != left {
			t.Fatalf("Base(left, right)=%s, want %s", first, left)
		}
	})
}

func TestQueryPeelsAnnotatedTags(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := memory.New(algo)
		blob := store.AddObject(objecttype.TypeBlob, []byte("blob\n"))
		leftTree := store.AddObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("left"),
			ID:   blob,
		}}}))
		rightTree := store.AddObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("right"),
			ID:   blob,
		}}}))
		base := store.AddObject(objecttype.TypeCommit, commitBody(leftTree))
		left := store.AddObject(objecttype.TypeCommit, commitBody(leftTree, base))
		right := store.AddObject(objecttype.TypeCommit, commitBody(rightTree, base))
		tag := store.AddObject(objecttype.TypeTag, tagBody(right, objecttype.TypeCommit))

		query := mergebase.Query(store, nil, left, tag)
		got := collectSeq(query.Seq())

		err := query.Err()
		if err != nil {
			t.Fatalf("query.Err(): %v", err)
		}

		if !slices.Equal(got, []objectid.ObjectID{base}) {
			t.Fatalf("Query(left, tag)=%v, want [%s]", got, base)
		}
	})
}

func TestQueryCrissCrossReturnsAllBestCommonAncestors(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := memory.New(algo)
		blob := store.AddObject(objecttype.TypeBlob, []byte("blob\n"))
		rootTree := store.AddObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("root"),
			ID:   blob,
		}}}))
		base1Tree := store.AddObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("base1"),
			ID:   blob,
		}}}))
		base2Tree := store.AddObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("base2"),
			ID:   blob,
		}}}))
		leftTree := store.AddObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("left"),
			ID:   blob,
		}}}))
		rightTree := store.AddObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("right"),
			ID:   blob,
		}}}))
		root := store.AddObject(objecttype.TypeCommit, commitBody(rootTree))
		base1 := store.AddObject(objecttype.TypeCommit, commitBody(base1Tree, root))
		base2 := store.AddObject(objecttype.TypeCommit, commitBody(base2Tree, root))
		left := store.AddObject(objecttype.TypeCommit, commitBody(leftTree, base1, base2))
		right := store.AddObject(objecttype.TypeCommit, commitBody(rightTree, base2, base1))

		query := mergebase.Query(store, nil, left, right)
		got := toSet(collectSeq(query.Seq()))

		err := query.Err()
		if err != nil {
			t.Fatalf("query.Err(): %v", err)
		}

		want := map[objectid.ObjectID]struct{}{base1: {}, base2: {}}
		if !maps.Equal(got, want) {
			t.Fatalf("Query(left, right)=%v, want %v", slices.Collect(maps.Keys(got)), slices.Collect(maps.Keys(want)))
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

func TestQueryReturnsNoResultWhenNoCommonAncestorExists(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := memory.New(algo)
		leftBlob := store.AddObject(objecttype.TypeBlob, []byte("left\n"))
		leftTree := store.AddObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("left"),
			ID:   leftBlob,
		}}}))
		rightBlob := store.AddObject(objecttype.TypeBlob, []byte("right\n"))
		rightTree := store.AddObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("right"),
			ID:   rightBlob,
		}}}))
		left := store.AddObject(objecttype.TypeCommit, commitBody(leftTree))
		right := store.AddObject(objecttype.TypeCommit, commitBody(rightTree))

		query := mergebase.Query(store, nil, left, right)
		got := collectSeq(query.Seq())

		err := query.Err()
		if err != nil {
			t.Fatalf("query.Err(): %v", err)
		}

		if len(got) != 0 {
			t.Fatalf("Query(left, right)=%v, want no results", got)
		}

		_, ok, err := mergebase.Base(store, nil, left, right)
		if err != nil {
			t.Fatalf("Base(left, right): %v", err)
		}

		if ok {
			t.Fatal("Base(left, right) unexpectedly reported a base")
		}
	})
}

func TestQueryRejectsNonCommitAfterPeel(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := memory.New(algo)
		blob := store.AddObject(objecttype.TypeBlob, []byte("blob\n"))
		tree := store.AddObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("f"),
			ID:   blob,
		}}}))
		commit := store.AddObject(objecttype.TypeCommit, commitBody(tree))
		tagToTree := store.AddObject(objecttype.TypeTag, tagBody(tree, objecttype.TypeTree))

		query := mergebase.Query(store, nil, commit, tagToTree)
		_ = collectSeq(query.Seq())

		err := query.Err()
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

func TestQuerySeqSingleUse(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := memory.New(algo)
		blob := store.AddObject(objecttype.TypeBlob, []byte("blob\n"))
		tree := store.AddObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("f"),
			ID:   blob,
		}}}))
		base := store.AddObject(objecttype.TypeCommit, commitBody(tree))
		left := store.AddObject(objecttype.TypeCommit, commitBody(tree, base))
		right := store.AddObject(objecttype.TypeCommit, commitBody(tree, left))

		query := mergebase.Query(store, nil, left, right)

		_ = collectSeq(query.Seq())
		again := collectSeq(query.Seq())

		if len(again) != 0 {
			t.Fatalf("second Seq() unexpectedly yielded %v", again)
		}

		err := query.Err()
		if err == nil {
			t.Fatal("expected error after second Seq()")
		}

		if err.Error() != "mergebase: sequence already consumed" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
