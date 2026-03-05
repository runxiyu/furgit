package ingest_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/format/pack/ingest"
	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/repository"
)

// thinBaseDistances enumerates candidate main-history distances for thin-pack
// probing.
var thinBaseDistances = [...]int{16, 24, 32, 48, 64, 96, 128, 160, 192, 224, 256, 320}

// pickThinBase selects one main-history base where git emits a truly thin pack
// for `head ^base`.
func pickThinBase(t *testing.T, sender *testgit.TestRepo, head objectid.ObjectID) objectid.ObjectID {
	t.Helper()

	for _, distance := range thinBaseDistances {
		base := sender.RevParse(t, fmt.Sprintf("refs/heads/main~%d", distance))
		revs := []string{head.String(), "^" + base.String()}
		if sender.PackObjectsIsThin(t, revs) {
			return base
		}
	}

	t.Fatalf("failed to find thin base for head %s", head.String())

	return objectid.ObjectID{}
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
		sender := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		sender.MakeManyObjectsHistory(t)
		head := sender.RevParse(t, "refs/heads/main")

		reader := sender.PackObjectsReader(t, []string{head.String()}, false)
		defer func() {
			err := reader.Close()
			if err != nil {
				t.Fatalf("close pack reader: %v", err)
			}
		}()

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

		result, err := ingest.Ingest(reader, packRoot, algo, false, true, nil)
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
		wantRaw := sender.Run(t, "rev-list", "--objects", "refs/heads/main")
		for line := range strings.SplitSeq(strings.TrimSpace(wantRaw), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) == 0 {
				continue
			}
			_ = receiver.Run(t, "cat-file", "-e", fields[0])
		}
	})
}

func TestIngestThinPackWithoutFixReturnsUnresolved(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		sender := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		sender.MakeManyObjectsHistory(t)
		sender.Repack(t, "-a", "-d", "-f", "--window=128", "--depth=128")

		head := sender.RevParse(t, "refs/heads/main")
		base := pickThinBase(t, sender, head)
		reader := sender.PackObjectsReader(t, []string{head.String(), "^" + base.String()}, true)
		defer func() {
			err := reader.Close()
			if err != nil {
				t.Fatalf("close pack reader: %v", err)
			}
		}()

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

		_, err = ingest.Ingest(reader, packRoot, algo, false, true, nil)
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
		sender := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		sender.MakeManyObjectsHistory(t)
		sender.Repack(t, "-a", "-d", "-f", "--window=128", "--depth=128")

		head := sender.RevParse(t, "refs/heads/main")
		base := pickThinBase(t, sender, head)
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

		baseReader := sender.PackObjectsReader(t, []string{base.String()}, false)
		_, err = ingest.Ingest(baseReader, packRoot, algo, false, false, nil)
		if err != nil {
			_ = baseReader.Close()
			t.Fatalf("ingest base pack: %v", err)
		}
		err = baseReader.Close()
		if err != nil {
			t.Fatalf("close base reader: %v", err)
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

		thinReader := sender.PackObjectsReader(t, []string{head.String(), "^" + base.String()}, true)
		result, err := ingest.Ingest(thinReader, packRoot, algo, true, true, receiverRepo.Objects())
		if err != nil {
			_ = thinReader.Close()
			t.Fatalf("Ingest(thin): %v", err)
		}
		err = thinReader.Close()
		if err != nil {
			t.Fatalf("close thin reader: %v", err)
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
		sender := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		sender.MakeManyObjectsHistory(t)
		head := sender.RevParse(t, "refs/heads/main")

		stream := sender.PackObjectsReader(t, []string{head.String()}, false)
		packBytes, err := io.ReadAll(stream)
		if err != nil {
			_ = stream.Close()
			t.Fatalf("read pack stream: %v", err)
		}
		err = stream.Close()
		if err != nil {
			t.Fatalf("close stream: %v", err)
		}
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
