package commitquery_test

import (
	"errors"
	"fmt"
	"testing"

	giterrors "codeberg.org/lindenii/furgit/errors"
	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore/memory"
	objecttype "codeberg.org/lindenii/furgit/object/type"

	"codeberg.org/lindenii/furgit/commitquery"
)

// ancestorCommitBody serializes one minimal commit body.
func ancestorCommitBody(tree objectid.ObjectID, parents ...objectid.ObjectID) []byte {
	buf := fmt.Appendf(nil, "tree %s\n", tree.String())
	for _, parent := range parents {
		buf = append(buf, fmt.Appendf(nil, "parent %s\n", parent.String())...)
	}

	buf = append(buf, []byte("\nmsg\n")...)

	return buf
}

// ancestorTagBody serializes one minimal annotated tag body.
func ancestorTagBody(target objectid.ObjectID, targetType objecttype.Type) []byte {
	targetName, ok := objecttype.Name(targetType)
	if !ok {
		panic("invalid tag target type")
	}

	return fmt.Appendf(nil, "object %s\ntype %s\ntag t\n\nmsg\n", target.String(), targetName)
}

// mustSerializeAncestorTree serializes one tree or fails the test.
func mustSerializeAncestorTree(tb testing.TB, tree *object.Tree) []byte {
	tb.Helper()

	body, err := tree.SerializeWithoutHeader()
	if err != nil {
		tb.Fatalf("SerializeWithoutHeader: %v", err)
	}

	return body
}

func TestIs(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := memory.New(algo)
		blob := store.AddObject(objecttype.TypeBlob, []byte("blob\n"))
		tree := store.AddObject(objecttype.TypeTree, mustSerializeAncestorTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("f"),
			ID:   blob,
		}}}))
		c1 := store.AddObject(objecttype.TypeCommit, ancestorCommitBody(tree))
		c2 := store.AddObject(objecttype.TypeCommit, ancestorCommitBody(tree, c1))
		otherBlob := store.AddObject(objecttype.TypeBlob, []byte("other-blob\n"))
		otherTree := store.AddObject(objecttype.TypeTree, mustSerializeAncestorTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("g"),
			ID:   otherBlob,
		}}}))
		c3 := store.AddObject(objecttype.TypeCommit, ancestorCommitBody(otherTree))
		tag := store.AddObject(objecttype.TypeTag, ancestorTagBody(c2, objecttype.TypeCommit))

		ok, err := commitquery.New(store, nil).IsAncestor(c1, tag)
		if err != nil {
			t.Fatalf("Is(c1, tag): %v", err)
		}

		if !ok {
			t.Fatal("expected c1 to be ancestor of tag->c2")
		}

		ok, err = commitquery.New(store, nil).IsAncestor(c3, c2)
		if err != nil {
			t.Fatalf("Is(c3, c2): %v", err)
		}

		if ok {
			t.Fatal("did not expect c3 to be ancestor of c2")
		}
	})
}

func TestIsRejectsNonCommitAfterPeel(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := memory.New(algo)
		blob := store.AddObject(objecttype.TypeBlob, []byte("blob\n"))
		tree := store.AddObject(objecttype.TypeTree, mustSerializeAncestorTree(t, &object.Tree{Entries: []object.TreeEntry{{
			Mode: object.FileModeRegular,
			Name: []byte("f"),
			ID:   blob,
		}}}))
		commit := store.AddObject(objecttype.TypeCommit, ancestorCommitBody(tree))
		tagToTree := store.AddObject(objecttype.TypeTag, ancestorTagBody(tree, objecttype.TypeTree))

		_, err := commitquery.New(store, nil).IsAncestor(commit, tagToTree)
		if err == nil {
			t.Fatal("expected error")
		}

		if _, ok := errors.AsType[*giterrors.ObjectTypeError](err); !ok {
			t.Fatalf("expected ObjectTypeError, got %T (%v)", err, err)
		}
	})
}
