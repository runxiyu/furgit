package loose_test

import (
	"io"
	"os"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	objectheader "codeberg.org/lindenii/furgit/object/header"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/store/loose"
	objecttype "codeberg.org/lindenii/furgit/object/type"
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

	header, ok := objectheader.Encode(ty, int64(len(body)))
	if !ok {
		t.Fatalf("objectheader.Encode failed")
	}

	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)

	return ty, body, raw
}

func corruptLooseObjectTrailer(t *testing.T, testRepo *testgit.TestRepo, id objectid.ObjectID) {
	t.Helper()

	root := testRepo.OpenObjectsRoot(t)

	hex := id.String()
	relPath := hex[:2] + "/" + hex[2:]

	file, err := root.OpenFile(relPath, os.O_RDWR, 0)
	if err != nil {
		t.Fatalf("OpenFile(%q): %v", relPath, err)
	}

	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		t.Fatalf("Stat(%q): %v", relPath, err)
	}

	if info.Size() == 0 {
		t.Fatalf("corrupt trailer on empty file %q", relPath)
	}

	last := make([]byte, 1)

	_, err = file.ReadAt(last, info.Size()-1)
	if err != nil {
		t.Fatalf("ReadAt(%q): %v", relPath, err)
	}

	last[0] ^= 0xff

	_, err = file.WriteAt(last, info.Size()-1)
	if err != nil {
		t.Fatalf("WriteAt(%q): %v", relPath, err)
	}
}
