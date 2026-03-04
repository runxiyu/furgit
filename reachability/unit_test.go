package reachability_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectheader"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	"codeberg.org/lindenii/furgit/objecttype"
	"codeberg.org/lindenii/furgit/reachability"
)

type storeObject struct {
	ty      objecttype.Type
	content []byte
}

type memStore struct {
	algo                objectid.Algorithm
	objects             map[objectid.ObjectID]storeObject
	readBytesByObjectID map[objectid.ObjectID]int
}

func newMemStore(algo objectid.Algorithm) *memStore {
	return &memStore{
		algo:                algo,
		objects:             make(map[objectid.ObjectID]storeObject),
		readBytesByObjectID: make(map[objectid.ObjectID]int),
	}
}

func (store *memStore) ReadBytesFull(id objectid.ObjectID) ([]byte, error) {
	obj, ok := store.objects[id]
	if !ok {
		return nil, objectstore.ErrObjectNotFound
	}

	header, ok := objectheader.Encode(obj.ty, int64(len(obj.content)))
	if !ok {
		panic("failed to encode object header")
	}

	raw := make([]byte, len(header)+len(obj.content))
	copy(raw, header)
	copy(raw[len(header):], obj.content)

	return raw, nil
}

func (store *memStore) ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error) {
	obj, ok := store.objects[id]
	if !ok {
		return objecttype.TypeInvalid, nil, objectstore.ErrObjectNotFound
	}

	store.readBytesByObjectID[id]++

	return obj.ty, append([]byte(nil), obj.content...), nil
}

func (store *memStore) ReadReaderFull(id objectid.ObjectID) (io.ReadCloser, error) {
	raw, err := store.ReadBytesFull(id)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(raw)), nil
}

func (store *memStore) ReadReaderContent(id objectid.ObjectID) (objecttype.Type, int64, io.ReadCloser, error) {
	ty, content, err := store.ReadBytesContent(id)
	if err != nil {
		return objecttype.TypeInvalid, 0, nil, err
	}

	return ty, int64(len(content)), io.NopCloser(bytes.NewReader(content)), nil
}

func (store *memStore) ReadSize(id objectid.ObjectID) (int64, error) {
	_, size, err := store.ReadHeader(id)
	if err != nil {
		return 0, err
	}

	return size, nil
}

func (store *memStore) ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error) {
	obj, ok := store.objects[id]
	if !ok {
		return objecttype.TypeInvalid, 0, objectstore.ErrObjectNotFound
	}

	return obj.ty, int64(len(obj.content)), nil
}

func (store *memStore) Close() error {
	return nil
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
		store := newMemStore(algo)
		blob := store.addObject(objecttype.TypeBlob, []byte("blob\n"))
		tree := store.addObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("f"),
			ID:   blob,
		}}}))
		commit1 := store.addObject(objecttype.TypeCommit, commitBody(tree))
		commit2 := store.addObject(objecttype.TypeCommit, commitBody(tree, commit1))
		tag1 := store.addObject(objecttype.TypeTag, tagBody(commit2, objecttype.TypeCommit))
		tag2 := store.addObject(objecttype.TypeTag, tagBody(tag1, objecttype.TypeTag))

		r := reachability.New(store)
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
		store := newMemStore(algo)
		blob := store.addObject(objecttype.TypeBlob, []byte("blob\n"))
		tree := store.addObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("f"),
			ID:   blob,
		}}}))
		commit := store.addObject(objecttype.TypeCommit, commitBody(tree))

		r := reachability.New(store)
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
		store := newMemStore(algo)
		blob := store.addObject(objecttype.TypeBlob, []byte("blob\n"))
		tree := store.addObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("f"),
			ID:   blob,
		}}}))
		tag := store.addObject(objecttype.TypeTag, tagBody(tree, objecttype.TypeTree))

		r := reachability.New(store)
		walk := r.Walk(reachability.DomainCommits, nil, map[objectid.ObjectID]struct{}{tag: {}})
		_ = collectSeq(walk.Seq())

		err := walk.Err()
		if err == nil {
			t.Fatal("expected error")
		}

		var typeErr *reachability.ErrObjectType
		if !errors.As(err, &typeErr) {
			t.Fatalf("expected ErrObjectType, got %T (%v)", err, err)
		}

		if typeErr.Got != objecttype.TypeTree || typeErr.Want != objecttype.TypeCommit {
			t.Fatalf("unexpected type error: %+v", typeErr)
		}
	})
}

func TestWalkDomainCommitsHaveTagStopsTraversal(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := newMemStore(algo)
		blob := store.addObject(objecttype.TypeBlob, []byte("blob\n"))
		tree := store.addObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("f"),
			ID:   blob,
		}}}))
		commit1 := store.addObject(objecttype.TypeCommit, commitBody(tree))
		commit2 := store.addObject(objecttype.TypeCommit, commitBody(tree, commit1))
		tag1 := store.addObject(objecttype.TypeTag, tagBody(commit2, objecttype.TypeCommit))
		tag2 := store.addObject(objecttype.TypeTag, tagBody(tag1, objecttype.TypeTag))

		r := reachability.New(store)
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
		store := newMemStore(algo)

		blob1 := store.addObject(objecttype.TypeBlob, []byte("b1\n"))
		blob2 := store.addObject(objecttype.TypeBlob, []byte("b2\n"))
		gitlinkTarget := store.algo.Sum([]byte("external-submodule"))

		subtree := store.addObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("nested"),
			ID:   blob2,
		}}}))
		rootTree := store.addObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{
			{Mode: object.FileModeRegular, Name: []byte("a"), ID: blob1},
			{Mode: object.FileModeDir, Name: []byte("dir"), ID: subtree},
			{Mode: object.FileModeGitlink, Name: []byte("submodule"), ID: gitlinkTarget},
		}}))
		commit := store.addObject(objecttype.TypeCommit, commitBody(rootTree))

		r := reachability.New(store)
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
		store := newMemStore(algo)
		blob := store.addObject(objecttype.TypeBlob, []byte("blob\n"))
		tree := store.addObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("f"),
			ID:   blob,
		}}}))
		missingParent := store.algo.Sum([]byte("missing-parent"))
		commit := store.addObject(objecttype.TypeCommit, commitBody(tree, missingParent))

		r := reachability.New(store)

		err := r.CheckConnected(reachability.DomainCommits, nil, map[objectid.ObjectID]struct{}{commit: {}})
		if err == nil {
			t.Fatal("expected error")
		}

		var missing *reachability.ErrObjectMissing
		if !errors.As(err, &missing) {
			t.Fatalf("expected ErrObjectMissing, got %T (%v)", err, err)
		}

		if missing.OID != missingParent {
			t.Fatalf("unexpected missing oid: got %s want %s", missing.OID, missingParent)
		}
	})
}

func TestWalkInvalidDomainReturnsPlainError(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		r := reachability.New(newMemStore(algo))
		walk := r.Walk(reachability.Domain(99), nil, nil)

		_ = collectSeq(walk.Seq())

		err := walk.Err()
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestIsAncestor(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := newMemStore(algo)
		blob := store.addObject(objecttype.TypeBlob, []byte("blob\n"))
		tree := store.addObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("f"),
			ID:   blob,
		}}}))
		c1 := store.addObject(objecttype.TypeCommit, commitBody(tree))
		c2 := store.addObject(objecttype.TypeCommit, commitBody(tree, c1))
		otherBlob := store.addObject(objecttype.TypeBlob, []byte("other-blob\n"))
		otherTree := store.addObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("g"),
			ID:   otherBlob,
		}}}))
		c3 := store.addObject(objecttype.TypeCommit, commitBody(otherTree))
		tag := store.addObject(objecttype.TypeTag, tagBody(c2, objecttype.TypeCommit))

		r := reachability.New(store)

		ok, err := r.IsAncestor(c1, tag)
		if err != nil {
			t.Fatalf("IsAncestor(c1, tag): %v", err)
		}

		if !ok {
			t.Fatal("expected c1 to be ancestor of tag->c2")
		}

		ok, err = r.IsAncestor(c3, c2)
		if err != nil {
			t.Fatalf("IsAncestor(c3, c2): %v", err)
		}

		if ok {
			t.Fatal("did not expect c3 to be ancestor of c2")
		}
	})
}

func TestIsAncestorRejectsNonCommitAfterPeel(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := newMemStore(algo)
		blob := store.addObject(objecttype.TypeBlob, []byte("blob\n"))
		tree := store.addObject(objecttype.TypeTree, mustSerializeTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("f"),
			ID:   blob,
		}}}))
		commit := store.addObject(objecttype.TypeCommit, commitBody(tree))
		tagToTree := store.addObject(objecttype.TypeTag, tagBody(tree, objecttype.TypeTree))

		r := reachability.New(store)

		_, err := r.IsAncestor(commit, tagToTree)
		if err == nil {
			t.Fatal("expected error")
		}

		var typeErr *reachability.ErrObjectType
		if !errors.As(err, &typeErr) {
			t.Fatalf("expected ErrObjectType, got %T (%v)", err, err)
		}
	})
}

func mustSerializeTree(tb testing.TB, tree *object.Tree) []byte {
	tb.Helper()

	body, err := tree.SerializeWithoutHeader()
	if err != nil {
		tb.Fatalf("SerializeWithoutHeader: %v", err)
	}

	return body
}

func (store *memStore) addObject(ty objecttype.Type, body []byte) objectid.ObjectID {
	header, ok := objectheader.Encode(ty, int64(len(body)))
	if !ok {
		panic("failed to encode object header")
	}

	raw := append(append([]byte(nil), header...), body...)
	id := store.algo.Sum(raw)
	store.objects[id] = storeObject{ty: ty, content: append([]byte(nil), body...)}

	return id
}
