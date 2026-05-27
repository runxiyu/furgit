package ingest_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	objectid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store/packed/internal/ingest"
)

type noExtraReadReader struct {
	reader *bytes.Reader
}

func (r *noExtraReadReader) Read(p []byte) (int, error) {
	if r.reader.Len() == 0 {
		return 0, errors.New("unexpected extra read after pack trailer")
	}

	return r.reader.Read(p)
}

// fixturePath returns one fixture file path for the selected algorithm.
func fixturePath(t *testing.T, algo objectid.Algorithm, name string) string {
	t.Helper()

	dir := algo.String()
	if dir == "" {
		t.Fatalf("unsupported fixture algorithm: %v", algo)
	}

	return filepath.Join("testdata", "fixtures", dir, name)
}

// fixtureBytes reads one fixture file fully.
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
func verifyReindexOracle(t *testing.T, repo *testgit.TestRepo, packName, idxName, revName string) {
	t.Helper()

	oracleDir := t.TempDir()
	oracleIdxPath := filepath.Join(oracleDir, "oracle.idx")
	_ = repo.Run(t, "index-pack", "--rev-index", "-o", oracleIdxPath, filepath.Join("objects", "pack", packName))
	oracleRevPath := strings.TrimSuffix(oracleIdxPath, ".idx") + ".rev"

	packRoot := repo.OpenPackRoot(t)

	gotIdx, err := packRoot.ReadFile(idxName)
	if err != nil {
		t.Fatalf("read idx: %v", err)
	}

	oracleRoot, err := os.OpenRoot(oracleDir)
	if err != nil {
		t.Fatalf("open oracle root: %v", err)
	}

	defer func() {
		err := oracleRoot.Close()
		if err != nil {
			t.Fatalf("close oracle root: %v", err)
		}
	}()

	wantIdx, err := oracleRoot.ReadFile(filepath.Base(oracleIdxPath))
	if err != nil {
		t.Fatalf("read oracle idx: %v", err)
	}

	if !bytes.Equal(gotIdx, wantIdx) {
		t.Fatal("idx bytes differ from git index-pack output")
	}

	gotRev, err := packRoot.ReadFile(revName)
	if err != nil {
		t.Fatalf("read rev: %v", err)
	}

	wantRev, err := oracleRoot.ReadFile(filepath.Base(oracleRevPath))
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

		packRoot := receiver.OpenPackRoot(t)

		result, err := ingest.WritePack(packRoot, algo, bytes.NewReader(packBytes), ingest.Options{
			WriteRev:           true,
			RequireTrailingEOF: true,
		})
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

		_ = receiver.Run(t, "verify-pack", "-v", filepath.Join("objects", "pack", result.IdxName))
		verifyReindexOracle(t, receiver, result.PackName, result.IdxName, result.RevName)

		receiver.UpdateRef(t, "refs/heads/main", head)
		_ = receiver.Run(t, "fsck", "--full", "--strict", "--no-progress", "--no-dangling")
	})
}

func TestIngestThinPackWithoutFixReturnsUnresolved(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		thinPack := fixtureBytes(t, algo, "thin.pack")

		receiver := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		packRoot := receiver.OpenPackRoot(t)

		_, err := ingest.WritePack(packRoot, algo, bytes.NewReader(thinPack), ingest.Options{
			WriteRev:           true,
			RequireTrailingEOF: true,
		})
		if err == nil {
			t.Fatal("Ingest error = nil, want error")
		}

		if _, ok := errors.AsType[*ingest.ThinPackUnresolvedError](err); !ok {
			t.Fatalf("Ingest error type = %T (%v), want *ThinPackUnresolvedError", err, err)
		}

		entries, err := fs.ReadDir(packRoot.FS(), ".")
		if err != nil {
			t.Fatalf("ReadDir(pack): %v", err)
		}

		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".pack") {
				t.Fatalf("found finalized pack file after failure: %v", entry.Name())
			}
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

		packRoot := receiver.OpenPackRoot(t)

		_, err := ingest.WritePack(packRoot, algo, bytes.NewReader(basePack), ingest.Options{
			RequireTrailingEOF: true,
		})
		if err != nil {
			t.Fatalf("ingest base pack: %v", err)
		}

		receiverRepo := receiver.OpenRepository(t)

		result, err := ingest.WritePack(packRoot, algo, bytes.NewReader(thinPack), ingest.Options{
			FixThin:            true,
			WriteRev:           true,
			ThinBase:           receiverRepo.ObjectStore(),
			RequireTrailingEOF: true,
		})
		if err != nil {
			t.Fatalf("Ingest(thin): %v", err)
		}

		if !result.ThinFixed {
			t.Fatal("ThinFixed = false, want true")
		}

		_ = receiver.Run(t, "verify-pack", "-v", filepath.Join("objects", "pack", result.IdxName))
		verifyReindexOracle(t, receiver, result.PackName, result.IdxName, result.RevName)
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
		packRoot := receiver.OpenPackRoot(t)

		_, err := ingest.WritePack(packRoot, algo, bytes.NewReader(packBytes), ingest.Options{
			WriteRev:           true,
			RequireTrailingEOF: true,
		})
		if err == nil {
			t.Fatal("Ingest error = nil, want error")
		}

		if _, ok := errors.AsType[*ingest.PackTrailerMismatchError](err); !ok {
			t.Fatalf("Ingest error type = %T (%v), want *PackTrailerMismatchError", err, err)
		}

		entries, err := fs.ReadDir(packRoot.FS(), ".")
		if err != nil {
			t.Fatalf("ReadDir(pack): %v", err)
		}

		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".pack") {
				t.Fatalf("found finalized pack file after failure: %v", entry.Name())
			}
		}
	})
}

func zeroObjectPackBytes(t *testing.T, algo objectid.Algorithm) []byte {
	t.Helper()

	hashImpl, err := algo.New()
	if err != nil {
		t.Fatalf("algo.New: %v", err)
	}

	var header [12]byte
	copy(header[:4], []byte{'P', 'A', 'C', 'K'})
	binary.BigEndian.PutUint32(header[4:8], 2)
	binary.BigEndian.PutUint32(header[8:12], 0)

	_, _ = hashImpl.Write(header[:])

	return append(header[:], hashImpl.Sum(nil)...)
}

func TestIngestZeroObjectPackIsDiscardedInternally(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		packBytes := zeroObjectPackBytes(t, algo)
		receiver := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		packRoot := receiver.OpenPackRoot(t)

		result, err := ingest.WritePack(packRoot, algo, bytes.NewReader(packBytes), ingest.Options{
			RequireTrailingEOF: true,
		})
		if err != nil {
			t.Fatalf("WritePack: %v", err)
		}

		if result.ObjectCount != 0 {
			t.Fatalf("ObjectCount = %d, want 0", result.ObjectCount)
		}

		if result.PackName != "" {
			t.Fatalf("PackName = %q, want empty", result.PackName)
		}

		if result.IdxName != "" {
			t.Fatalf("IdxName = %q, want empty", result.IdxName)
		}

		if result.RevName != "" {
			t.Fatalf("RevName = %q, want empty", result.RevName)
		}

		entries, err := fs.ReadDir(packRoot.FS(), ".")
		if err != nil {
			t.Fatalf("ReadDir(pack): %v", err)
		}

		if len(entries) != 0 {
			t.Fatalf("unexpected files after zero-object pack: %d", len(entries))
		}
	})
}

func TestIngestCanFinishWithoutTrailingEOF(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		head := fixtureOID(t, algo, "head")
		packBytes := fixtureBytes(t, algo, "nonthin.pack")

		receiver := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		packRoot := receiver.OpenPackRoot(t)

		result, err := ingest.WritePack(packRoot, algo, &noExtraReadReader{reader: bytes.NewReader(packBytes)}, ingest.Options{
			WriteRev: true,
		})
		if err != nil {
			t.Fatalf("Ingest without trailing EOF: %v", err)
		}

		receiver.UpdateRef(t, "refs/heads/main", head)
		_ = receiver.Run(t, "verify-pack", "-v", filepath.Join("objects", "pack", result.IdxName))
		_ = receiver.Run(t, "fsck", "--full", "--strict", "--no-progress", "--no-dangling")
	})
}
