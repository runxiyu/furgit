package ingest_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/format/pack/ingest"
	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/repository"
)

func fixtureAlgorithmDir(algo objectid.Algorithm) string {
	switch algo { //nolint:exhaustive
	case objectid.AlgorithmSHA1:
		return "sha1"
	case objectid.AlgorithmSHA256:
		return "sha256"
	default:
		return ""
	}
}

// fixturePath returns one fixture file path for the selected algorithm.
func fixturePath(t *testing.T, algo objectid.Algorithm, name string) string {
	t.Helper()

	dir := fixtureAlgorithmDir(algo)
	if dir == "" {
		t.Fatalf("unsupported fixture algorithm: %v", algo)
	}

	return filepath.Join("testdata", "fixtures", dir, name)
}

// fixtureBytes reads one fixture file fully.
func fixtureBytes(t *testing.T, algo objectid.Algorithm, name string) []byte {
	t.Helper()

	path := fixturePath(t, algo, name)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %q: %v", path, err)
	}

	return data
}

// fixtureMetadata parses key=value metadata for one algorithm fixture set.
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

// fixtureOID returns one fixture metadata object ID value.
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

// verifyReindexOracle regenerates idx/rev with upstream git index-pack and
// compares bytes with files produced by ingest.
func verifyReindexOracle(t *testing.T, repo *testgit.TestRepo, packPath, idxPath, revPath string) {
	t.Helper()

	oracleDir := t.TempDir()
	oracleIdxPath := filepath.Join(oracleDir, "oracle.idx")
	_ = repo.Run(t, "index-pack", "--rev-index", "-o", oracleIdxPath, packPath)
	oracleRevPath := strings.TrimSuffix(oracleIdxPath, ".idx") + ".rev"

	gotIdx, err := os.ReadFile(idxPath)
	if err != nil {
		t.Fatalf("read idx: %v", err)
	}

	wantIdx, err := os.ReadFile(oracleIdxPath)
	if err != nil {
		t.Fatalf("read oracle idx: %v", err)
	}

	if !bytes.Equal(gotIdx, wantIdx) {
		t.Fatal("idx bytes differ from git index-pack output")
	}

	gotRev, err := os.ReadFile(revPath)
	if err != nil {
		t.Fatalf("read rev: %v", err)
	}

	wantRev, err := os.ReadFile(oracleRevPath)
	if err != nil {
		t.Fatalf("read oracle rev: %v", err)
	}

	if !bytes.Equal(gotRev, wantRev) {
		t.Fatal("rev bytes differ from git index-pack output")
	}
}

func TestIngestNonThinPackWritesPackIdxRev(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		head := fixtureOID(t, algo, "head")
		packBytes := fixtureBytes(t, algo, "nonthin.pack")

		receiver := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})

		packRoot, err := os.OpenRoot(filepath.Join(receiver.Dir(), "objects", "pack"))
		if err != nil {
			t.Fatalf("open pack root: %v", err)
		}

		defer func() {
			err = packRoot.Close()
			if err != nil {
				t.Fatalf("close pack root: %v", err)
			}
		}()

		result, err := ingest.Ingest(bytes.NewReader(packBytes), packRoot, algo, false, true, nil)
		if err != nil {
			t.Fatalf("Ingest: %v", err)
		}

		if result.ThinFixed {
			t.Fatalf("ThinFixed = true, want false")
		}

		if result.RevName == "" {
			t.Fatal("RevName is empty")
		}

		_, err = packRoot.Stat(result.PackName)
		if err != nil {
			t.Fatalf("stat pack: %v", err)
		}

		_, err = packRoot.Stat(result.IdxName)
		if err != nil {
			t.Fatalf("stat idx: %v", err)
		}

		_, err = packRoot.Stat(result.RevName)
		if err != nil {
			t.Fatalf("stat rev: %v", err)
		}

		idxPath := filepath.Join(receiver.Dir(), "objects", "pack", result.IdxName)
		packPath := filepath.Join(receiver.Dir(), "objects", "pack", result.PackName)
		revPath := filepath.Join(receiver.Dir(), "objects", "pack", result.RevName)
		_ = receiver.Run(t, "verify-pack", "-v", idxPath)
		verifyReindexOracle(t, receiver, packPath, idxPath, revPath)

		receiver.UpdateRef(t, "refs/heads/main", head)
		_ = receiver.Run(t, "fsck", "--full", "--strict", "--no-progress", "--no-dangling")
	})
}

func TestIngestThinPackWithoutFixReturnsUnresolved(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		thinPack := fixtureBytes(t, algo, "thin.pack")

		receiver := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		packDir := filepath.Join(receiver.Dir(), "objects", "pack")

		packRoot, err := os.OpenRoot(packDir)
		if err != nil {
			t.Fatalf("open pack root: %v", err)
		}

		defer func() {
			err = packRoot.Close()
			if err != nil {
				t.Fatalf("close pack root: %v", err)
			}
		}()

		_, err = ingest.Ingest(bytes.NewReader(thinPack), packRoot, algo, false, true, nil)
		if err == nil {
			t.Fatal("Ingest error = nil, want error")
		}

		var unresolved *ingest.ErrThinPackUnresolved
		if !errors.As(err, &unresolved) {
			t.Fatalf("Ingest error type = %T (%v), want *ErrThinPackUnresolved", err, err)
		}

		matches, err := filepath.Glob(filepath.Join(packDir, "pack-*.pack"))
		if err != nil {
			t.Fatalf("glob pack files: %v", err)
		}

		if len(matches) != 0 {
			t.Fatalf("found finalized pack files after failure: %v", matches)
		}
	})
}

func TestIngestThinPackWithFixThin(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		head := fixtureOID(t, algo, "head")
		basePack := fixtureBytes(t, algo, "base.pack")
		thinPack := fixtureBytes(t, algo, "thin.pack")
		receiver := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})

		packRoot, err := os.OpenRoot(filepath.Join(receiver.Dir(), "objects", "pack"))
		if err != nil {
			t.Fatalf("open pack root: %v", err)
		}

		defer func() {
			err = packRoot.Close()
			if err != nil {
				t.Fatalf("close pack root: %v", err)
			}
		}()

		_, err = ingest.Ingest(bytes.NewReader(basePack), packRoot, algo, false, false, nil)
		if err != nil {
			t.Fatalf("ingest base pack: %v", err)
		}

		receiverRoot, err := os.OpenRoot(receiver.Dir())
		if err != nil {
			t.Fatalf("open receiver root: %v", err)
		}

		defer func() {
			err = receiverRoot.Close()
			if err != nil {
				t.Fatalf("close receiver root: %v", err)
			}
		}()

		receiverRepo, err := repository.Open(receiverRoot)
		if err != nil {
			t.Fatalf("repository.Open(receiver): %v", err)
		}

		defer func() {
			err = receiverRepo.Close()
			if err != nil {
				t.Fatalf("close receiver repo: %v", err)
			}
		}()

		result, err := ingest.Ingest(bytes.NewReader(thinPack), packRoot, algo, true, true, receiverRepo.Objects())
		if err != nil {
			t.Fatalf("Ingest(thin): %v", err)
		}

		if !result.ThinFixed {
			t.Fatal("ThinFixed = false, want true")
		}

		idxPath := filepath.Join(receiver.Dir(), "objects", "pack", result.IdxName)
		packPath := filepath.Join(receiver.Dir(), "objects", "pack", result.PackName)
		revPath := filepath.Join(receiver.Dir(), "objects", "pack", result.RevName)
		_ = receiver.Run(t, "verify-pack", "-v", idxPath)
		verifyReindexOracle(t, receiver, packPath, idxPath, revPath)
		receiver.UpdateRef(t, "refs/heads/main", head)
		_ = receiver.Run(t, "fsck", "--full", "--strict", "--no-progress", "--no-dangling")
	})
}

func TestIngestPackTrailerMismatch(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		packBytes := fixtureBytes(t, algo, "nonthin.pack")
		if len(packBytes) == 0 {
			t.Fatal("empty pack stream")
		}

		packBytes[len(packBytes)-1] ^= 0xff

		receiver := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		packDir := filepath.Join(receiver.Dir(), "objects", "pack")

		packRoot, err := os.OpenRoot(packDir)
		if err != nil {
			t.Fatalf("open pack root: %v", err)
		}

		defer func() {
			err = packRoot.Close()
			if err != nil {
				t.Fatalf("close pack root: %v", err)
			}
		}()

		_, err = ingest.Ingest(bytes.NewReader(packBytes), packRoot, algo, false, true, nil)
		if err == nil {
			t.Fatal("Ingest error = nil, want error")
		}

		var mismatch *ingest.ErrPackTrailerMismatch
		if !errors.As(err, &mismatch) {
			t.Fatalf("Ingest error type = %T (%v), want *ErrPackTrailerMismatch", err, err)
		}

		matches, err := filepath.Glob(filepath.Join(packDir, "pack-*.pack"))
		if err != nil {
			t.Fatalf("glob pack files: %v", err)
		}

		if len(matches) != 0 {
			t.Fatalf("found finalized pack files after failure: %v", matches)
		}
	})
}
