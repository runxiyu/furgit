package packed_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstore "codeberg.org/lindenii/furgit/object/store"
	"codeberg.org/lindenii/furgit/object/store/packed"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

func fixturePath(t *testing.T, algo objectid.Algorithm, name string) string {
	t.Helper()

	return filepath.Join("internal", "ingest", "testdata", "fixtures", algo.String(), name)
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

func TestPackQuarantinePromotePublishesWrittenObjects(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		head := fixtureOID(t, algo, "head")
		packBytes := fixtureBytes(t, algo, "nonthin.pack")

		repo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		packRoot := repo.OpenPackRoot(t)

		store, err := packed.New(packRoot, algo, packed.Options{WriteRev: true})
		if err != nil {
			t.Fatalf("packed.New: %v", err)
		}

		defer func() {
			err := store.Close()
			if err != nil {
				t.Fatalf("store.Close: %v", err)
			}
		}()

		quarantiner, ok := any(store).(objectstore.PackQuarantiner)
		if !ok {
			t.Fatal("packed store does not implement PackQuarantiner")
		}

		quarantine, err := quarantiner.BeginPackQuarantine(objectstore.PackQuarantineOptions{})
		if err != nil {
			t.Fatalf("BeginPackQuarantine: %v", err)
		}

		err = quarantine.WritePack(bytes.NewReader(packBytes), objectstore.PackWriteOptions{RequireTrailingEOF: true})
		if err != nil {
			t.Fatalf("quarantine.WritePack: %v", err)
		}

		ty, _, err := quarantine.ReadHeader(head)
		if err != nil {
			t.Fatalf("quarantine.ReadHeader: %v", err)
		}

		if ty != objecttype.TypeCommit {
			t.Fatalf("quarantine.ReadHeader type = %v, want commit", ty)
		}

		_, _, err = store.ReadHeader(head)
		if err == nil {
			t.Fatal("store.ReadHeader unexpectedly saw quarantined object before promote")
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
			t.Fatalf("store.ReadHeader after promote: %v", err)
		}

		if ty != objecttype.TypeCommit {
			t.Fatalf("store.ReadHeader type = %v, want commit", ty)
		}
	})
}

func TestPackQuarantineDiscardDropsWrittenObjects(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		head := fixtureOID(t, algo, "head")
		packBytes := fixtureBytes(t, algo, "nonthin.pack")

		repo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		packRoot := repo.OpenPackRoot(t)

		store, err := packed.New(packRoot, algo, packed.Options{WriteRev: true})
		if err != nil {
			t.Fatalf("packed.New: %v", err)
		}

		defer func() {
			err := store.Close()
			if err != nil {
				t.Fatalf("store.Close: %v", err)
			}
		}()

		quarantiner, ok := any(store).(objectstore.PackQuarantiner)
		if !ok {
			t.Fatalf("expected objectstore.PackQuarantiner")
		}

		quarantine, err := quarantiner.BeginPackQuarantine(objectstore.PackQuarantineOptions{})
		if err != nil {
			t.Fatalf("BeginPackQuarantine: %v", err)
		}

		err = quarantine.WritePack(bytes.NewReader(packBytes), objectstore.PackWriteOptions{RequireTrailingEOF: true})
		if err != nil {
			t.Fatalf("quarantine.WritePack: %v", err)
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
			t.Fatal("store.ReadHeader unexpectedly saw discarded object")
		}
	})
}
