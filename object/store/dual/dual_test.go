package dual_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstore "codeberg.org/lindenii/furgit/object/store"
	"codeberg.org/lindenii/furgit/object/store/dual"
	"codeberg.org/lindenii/furgit/object/store/loose"
	"codeberg.org/lindenii/furgit/object/store/packed"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

func fixturePath(t *testing.T, algo objectid.Algorithm, name string) string {
	t.Helper()

	return filepath.Join("..", "packed", "internal", "ingest", "testdata", "fixtures", algo.String(), name)
}

func fixtureBytes(t *testing.T, algo objectid.Algorithm, name string) []byte {
	t.Helper()

	path := fixturePath(t, algo, name)
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatalf("open fixture root %q: %v", dir, err)
	}

	defer func() {
		err := root.Close()
		if err != nil {
			t.Fatalf("close fixture root %q: %v", dir, err)
		}
	}()

	data, err := root.ReadFile(base)
	if err != nil {
		t.Fatalf("read fixture %q: %v", base, err)
	}

	return data
}

func fixtureMetadata(t *testing.T, algo objectid.Algorithm) map[string]string {
	t.Helper()

	data := fixtureBytes(t, algo, "METADATA.txt")
	out := make(map[string]string)

	for line := range strings.SplitSeq(strings.TrimSpace(string(data)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			t.Fatalf("invalid fixture metadata line %q", line)
		}

		out[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}

	return out
}

func fixtureOID(t *testing.T, algo objectid.Algorithm, key string) objectid.ObjectID {
	t.Helper()

	meta := fixtureMetadata(t, algo)
	hex, ok := meta[key]
	if !ok {
		t.Fatalf("missing fixture metadata key %q", key)
	}

	id, err := objectid.ParseHex(algo, hex)
	if err != nil {
		t.Fatalf("parse fixture metadata oid %q: %v", hex, err)
	}

	return id
}

func newDualStore(t *testing.T, repo *testgit.TestRepo, algo objectid.Algorithm) *dual.Dual {
	t.Helper()

	objectsRoot := repo.OpenObjectsRoot(t)
	looseStore, err := loose.New(objectsRoot, algo)
	if err != nil {
		t.Fatalf("loose.New: %v", err)
	}

	packRoot := repo.OpenPackRoot(t)
	packedStore, err := packed.New(packRoot, algo, packed.Options{WriteRev: true})
	if err != nil {
		t.Fatalf("packed.New: %v", err)
	}

	return dual.New(looseStore, packedStore)
}

func TestDualReadsWritesAndQuarantine(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		head := fixtureOID(t, algo, "head")
		packBytes := fixtureBytes(t, algo, "nonthin.pack")

		repo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		store := newDualStore(t, repo, algo)

		quarantiner, ok := any(store).(objectstore.Quarantiner)
		if !ok {
			t.Fatal("dual does not implement Quarantiner")
		}

		quarantine, err := quarantiner.BeginQuarantine(objectstore.QuarantineOptions{})
		if err != nil {
			t.Fatalf("BeginQuarantine: %v", err)
		}

		err = quarantine.WritePack(bytes.NewReader(packBytes), objectstore.PackWriteOptions{RequireTrailingEOF: true})
		if err != nil {
			t.Fatalf("quarantine.WritePack: %v", err)
		}

		objectQ, ok := any(quarantine).(objectstore.ObjectQuarantine)
		if !ok {
			t.Fatal("pack quarantine does not also implement ObjectQuarantine")
		}

		looseContent := []byte("dual quarantine loose object\n")
		looseID, err := objectQ.WriteBytesContent(objecttype.TypeBlob, looseContent)
		if err != nil {
			t.Fatalf("quarantine.WriteBytesContent: %v", err)
		}

		ty, _, err := quarantine.ReadHeader(head)
		if err != nil {
			t.Fatalf("quarantine.ReadHeader(pack): %v", err)
		}

		if ty != objecttype.TypeCommit {
			t.Fatalf("quarantine.ReadHeader(pack) type = %v, want commit", ty)
		}

		ty, got, err := quarantine.ReadBytesContent(looseID)
		if err != nil {
			t.Fatalf("quarantine.ReadBytesContent(loose): %v", err)
		}

		if ty != objecttype.TypeBlob {
			t.Fatalf("quarantine.ReadBytesContent(loose) type = %v, want blob", ty)
		}

		if !bytes.Equal(got, looseContent) {
			t.Fatal("quarantine.ReadBytesContent(loose) mismatch")
		}

		_, _, err = store.ReadHeader(head)
		if err == nil {
			t.Fatal("store.ReadHeader unexpectedly saw quarantined pack object before promote")
		}

		_, _, err = store.ReadBytesContent(looseID)
		if err == nil {
			t.Fatal("store.ReadBytesContent unexpectedly saw quarantined loose object before promote")
		}

		err = quarantine.Promote()
		if err != nil {
			t.Fatalf("quarantine.Promote: %v", err)
		}

		err = store.Refresh()
		if err != nil {
			t.Fatalf("store.Refresh: %v", err)
		}

		ty, _, err = store.ReadHeader(head)
		if err != nil {
			t.Fatalf("store.ReadHeader(pack): %v", err)
		}

		if ty != objecttype.TypeCommit {
			t.Fatalf("store.ReadHeader(pack) type = %v, want commit", ty)
		}

		ty, got, err = store.ReadBytesContent(looseID)
		if err != nil {
			t.Fatalf("store.ReadBytesContent(loose): %v", err)
		}

		if ty != objecttype.TypeBlob {
			t.Fatalf("store.ReadBytesContent(loose) type = %v, want blob", ty)
		}

		if !bytes.Equal(got, looseContent) {
			t.Fatal("store.ReadBytesContent(loose) mismatch")
		}
	})
}

func TestDualQuarantineDiscardDropsBothHalves(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		head := fixtureOID(t, algo, "head")
		packBytes := fixtureBytes(t, algo, "nonthin.pack")

		repo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		store := newDualStore(t, repo, algo)

		quarantiner := any(store).(objectstore.Quarantiner)
		quarantine, err := quarantiner.BeginQuarantine(objectstore.QuarantineOptions{})
		if err != nil {
			t.Fatalf("BeginQuarantine: %v", err)
		}

		err = quarantine.WritePack(bytes.NewReader(packBytes), objectstore.PackWriteOptions{RequireTrailingEOF: true})
		if err != nil {
			t.Fatalf("quarantine.WritePack: %v", err)
		}

		looseID, err := quarantine.WriteBytesContent(objecttype.TypeBlob, []byte("discarded dual object\n"))
		if err != nil {
			t.Fatalf("quarantine.WriteBytesContent: %v", err)
		}

		err = quarantine.Discard()
		if err != nil {
			t.Fatalf("quarantine.Discard: %v", err)
		}

		err = store.Refresh()
		if err != nil {
			t.Fatalf("store.Refresh: %v", err)
		}

		_, _, err = store.ReadHeader(head)
		if err == nil {
			t.Fatal("store.ReadHeader unexpectedly saw discarded pack object")
		}

		_, _, err = store.ReadBytesContent(looseID)
		if err == nil {
			t.Fatal("store.ReadBytesContent unexpectedly saw discarded loose object")
		}
	})
}
