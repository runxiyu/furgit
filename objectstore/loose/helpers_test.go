package loose_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectheader"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore/loose"
	"codeberg.org/lindenii/furgit/objecttype"
)

func openLooseStore(t *testing.T, repoPath string, algo objectid.Algorithm) *loose.Store {
	t.Helper()
	objectsPath := filepath.Join(repoPath, "objects")
	root, err := os.OpenRoot(objectsPath)
	if err != nil {
		t.Fatalf("OpenRoot(%q): %v", objectsPath, err)
	}
	t.Cleanup(func() { _ = root.Close() })

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
	if err := reader.Close(); err != nil {
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
