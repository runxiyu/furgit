package commitquery_test

import (
	"errors"
	"fmt"
	"testing"

	giterrors "lindenii.org/go/furgit/errors"
	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/fetch"
	objectid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store/memory"
	objecttree "lindenii.org/go/furgit/object/tree"
	objecttype "lindenii.org/go/furgit/object/type"

	"lindenii.org/go/furgit/commitquery"
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
	targetName, ok := targetType.Name()
	if !ok {
		panic("invalid tag target type")
	}

	return fmt.Appendf(nil, "object %s\ntype %s\ntag t\n\nmsg\n", target.String(), targetName)
}

// mustSerializeAncestorTree serializes one tree or fails the test.
func mustSerializeAncestorTree(tb testing.TB, tree *objecttree.Tree) []byte {
	tb.Helper()

	body, err := tree.BytesWithoutHeader()
	if err != nil {
		tb.Fatalf("BytesWithoutHeader: %v", err)
	}

	return body
}

func TestIs(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		store := memory.New(algo)

		blob, err := store.WriteBytesContent(objecttype.TypeBlob, []byte("blob\n"))
		if err != nil {
			t.Fatal(err)
		}

		tree, err := store.WriteBytesContent(objecttype.TypeTree, mustSerializeAncestorTree(t, &objecttree.Tree{Entries: []objecttree.TreeEntry{{
			Mode: objecttree.FileModeRegular,
			Name: []byte("f"),
			ID:   blob,
		}}}))
		if err != nil {
			t.Fatal(err)
		}

		c1, err := store.WriteBytesContent(objecttype.TypeCommit, ancestorCommitBody(tree))
		if err != nil {
			t.Fatal(err)
		}

		c2, err := store.WriteBytesContent(objecttype.TypeCommit, ancestorCommitBody(tree, c1))
		if err != nil {
			t.Fatal(err)
		}

		otherBlob, err := store.WriteBytesContent(objecttype.TypeBlob, []byte("other-blob\n"))
		if err != nil {
			t.Fatal(err)
		}

		otherTree, err := store.WriteBytesContent(objecttype.TypeTree, mustSerializeAncestorTree(t, &objecttree.Tree{Entries: []objecttree.TreeEntry{{
			Mode: objecttree.FileModeRegular,
			Name: []byte("g"),
			ID:   otherBlob,
		}}}))
		if err != nil {
			t.Fatal(err)
		}

		c3, err := store.WriteBytesContent(objecttype.TypeCommit, ancestorCommitBody(otherTree))
		if err != nil {
			t.Fatal(err)
		}

		tag, err := store.WriteBytesContent(objecttype.TypeTag, ancestorTagBody(c2, objecttype.TypeCommit))
		if err != nil {
			t.Fatal(err)
		}

		ok, err := commitquery.New(fetch.New(store), nil).IsAncestor(c1, tag)
		if err != nil {
			t.Fatalf("Is(c1, tag): %v", err)
		}

		if !ok {
			t.Fatal("expected c1 to be ancestor of tag->c2")
		}

		ok, err = commitquery.New(fetch.New(store), nil).IsAncestor(c3, c2)
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

		blob, err := store.WriteBytesContent(objecttype.TypeBlob, []byte("blob\n"))
		if err != nil {
			t.Fatal(err)
		}

		tree, err := store.WriteBytesContent(objecttype.TypeTree, mustSerializeAncestorTree(t, &objecttree.Tree{Entries: []objecttree.TreeEntry{{
			Mode: objecttree.FileModeRegular,
			Name: []byte("f"),
			ID:   blob,
		}}}))
		if err != nil {
			t.Fatal(err)
		}

		commit, err := store.WriteBytesContent(objecttype.TypeCommit, ancestorCommitBody(tree))
		if err != nil {
			t.Fatal(err)
		}

		tagToTree, err := store.WriteBytesContent(objecttype.TypeTag, ancestorTagBody(tree, objecttype.TypeTree))
		if err != nil {
			t.Fatal(err)
		}

		_, err = commitquery.New(fetch.New(store), nil).IsAncestor(commit, tagToTree)
		if err == nil {
			t.Fatal("expected error")
		}

		if _, ok := errors.AsType[*giterrors.ObjectTypeError](err); !ok {
			t.Fatalf("expected ObjectTypeError, got %T (%v)", err, err)
		}
	})
}
