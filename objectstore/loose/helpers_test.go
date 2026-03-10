package loose_test

import (
	"io"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object/header"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore/loose"
	"codeberg.org/lindenii/furgit/objecttype"
)

func openLooseStore(t *testing.T, testRepo *testgit.TestRepo, algo objectid.Algorithm) *loose.Store {
	t.Helper()

	root := testRepo.OpenObjectsRoot(t)

	store, err := loose.New(root, algo)
	if err != nil {
		t.Fatalf("loose.New: %v", err)
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

	header, ok := header.Encode(ty, int64(len(body)))
	if !ok {
		t.Fatalf("header.Encode failed")
	}

	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)

	return ty, body, raw
}
