package packed_test

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	objectheader "codeberg.org/lindenii/furgit/object/header"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/store/packed"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

func openPackedStore(t *testing.T, testRepo *testgit.TestRepo, algo objectid.Algorithm) *packed.Store {
	t.Helper()

	root := testRepo.OpenPackRoot(t)

	store, err := packed.New(root, algo, packed.Options{})
	if err != nil {
		t.Fatalf("packed.New: %v", err)
	}

	return store
}

func mustReadAllAndClose(t *testing.T, reader io.ReadCloser) []byte {
	t.Helper()

	data, err := io.ReadAll(reader)
	if err != nil {
		_ = reader.Close()

		t.Fatalf("ReadAll: %v", err)
	}

	err = reader.Close()
	if err != nil {
		t.Fatalf("Close: %v", err)
	}

	return data
}

func expectedRawObject(t *testing.T, testRepo *testgit.TestRepo, id objectid.ObjectID) (objecttype.Type, []byte, []byte) {
	t.Helper()

	typeName := testRepo.Run(t, "cat-file", "-t", id.String())

	ty, ok := objecttype.ParseName(typeName)
	if !ok {
		t.Fatalf("ParseName(%q) failed", typeName)
	}

	body := testRepo.CatFile(t, typeName, id)

	header, ok := objectheader.Encode(ty, int64(len(body)))
	if !ok {
		t.Fatalf("objectheader.Encode failed")
	}

	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)

	return ty, body, raw
}

func createPackedFixtureRepo(t *testing.T, algo objectid.Algorithm) (*testgit.TestRepo, []objectid.ObjectID) {
	t.Helper()

	testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
	blobID, treeID, commitID := testRepo.MakeCommit(t, "packed store base commit")
	testRepo.Run(t, "update-ref", "refs/heads/main", commitID.String())
	tagID := testRepo.TagAnnotated(t, "v1.0.0", commitID, "packed-store-tag")

	parent := commitID

	for i := range 24 {
		content := "common-prefix\n" + strings.Repeat("line-"+strconv.Itoa(i%3)+"\n", 256) + fmt.Sprintf("tail-%d\n", i)
		nextBlob, nextTree := testRepo.MakeSingleFileTree(t, fmt.Sprintf("file-%02d.txt", i), []byte(content))
		nextCommit := testRepo.CommitTree(t, nextTree, fmt.Sprintf("commit-%02d", i), parent)
		testRepo.Run(t, "update-ref", "refs/heads/main", nextCommit.String())
		parent = nextCommit

		_ = nextBlob
		_ = nextTree
	}

	testRepo.Repack(t, "-a", "-d", "-f", "--window=64", "--depth=64")

	return testRepo, []objectid.ObjectID{
		blobID,
		treeID,
		commitID,
		tagID,
		parent,
	}
}
