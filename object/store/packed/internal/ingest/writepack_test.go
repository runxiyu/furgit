package ingest_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/store/packed"
	"lindenii.org/go/furgit/object/store/packed/internal/ingest"
)

// TestWritePackMatchesGit verifies that ingesting a normal pack
// matches git's own pack, index, and reverse index.
//
// The pack is streamed through verbatim,
// and the index and reverse index are regenerated deterministically,
// so a successful match also confirms that scanning and delta resolution
// recovered every object ID, offset, and CRC that git recorded.
func TestWritePackMatchesGit(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			seeded, err := repo.SeedHistory(t)
			if err != nil {
				t.Fatalf("SeedHistory: %v", err)
			}

			gitPrefix, err := repo.PackObjects(t, seeded.All(), testgit.PackObjectsOptions{
				RevIndex: true,
				Revs:     false,
				Exclude:  nil,
			})
			if err != nil {
				t.Fatalf("PackObjects: %v", err)
			}

			stream, err := os.ReadFile(gitPrefix + ".pack") //nolint:gosec
			if err != nil {
				t.Fatalf("ReadFile pack: %v", err)
			}

			dir, result := writePack(t, objectFormat, bytes.NewReader(stream), store.PackWriteOptions{
				ThinBase: nil,
				Progress: nil,
			})

			if want := filepath.Base(gitPrefix) + ".pack"; result.PackName != want {
				t.Fatalf("PackName = %q, want %q", result.PackName, want)
			}

			for _, artifact := range []struct {
				kind string
				ours string
				want string
			}{
				{"pack", result.PackName, gitPrefix + ".pack"},
				{"idx", result.IdxName, gitPrefix + ".idx"},
				{"rev", result.RevName, gitPrefix + ".rev"},
			} {
				ours, err := os.ReadFile(filepath.Join(dir, artifact.ours)) //nolint:gosec
				if err != nil {
					t.Fatalf("ReadFile %s: %v", artifact.kind, err)
				}

				want, err := os.ReadFile(artifact.want)
				if err != nil {
					t.Fatalf("ReadFile git %s: %v", artifact.kind, err)
				}

				if !bytes.Equal(ours, want) {
					t.Errorf("%s differs from git: %d bytes vs %d", artifact.kind, len(ours), len(want))
				}
			}
		})
	}
}

// TestWritePackEmpty verifies that a zero-object pack
// succeeds without writing any artifacts.
func TestWritePackEmpty(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			stream, err := repo.PackObjectsStdout(t, nil, testgit.PackObjectsStdoutOptions{
				Revs:    false,
				Thin:    false,
				Exclude: nil,
			})
			if err != nil {
				t.Fatalf("PackObjectsStdout: %v", err)
			}

			dir, result := writePack(t, objectFormat, bytes.NewReader(stream), store.PackWriteOptions{
				ThinBase: nil,
				Progress: nil,
			})

			if result.ObjectCount != 0 {
				t.Fatalf("ObjectCount = %d, want 0", result.ObjectCount)
			}

			if result.PackName != "" {
				t.Fatalf("PackName = %q, want empty", result.PackName)
			}

			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatalf("ReadDir: %v", err)
			}

			if len(entries) != 0 {
				t.Fatalf("empty pack wrote %d files, want 0", len(entries))
			}
		})
	}
}

// TestWritePackIdempotent verifies that ingesting the same pack twice
// into one store succeeds and leaves the artifacts in place.
func TestWritePackIdempotent(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			seeded, err := repo.SeedHistory(t)
			if err != nil {
				t.Fatalf("SeedHistory: %v", err)
			}

			gitPrefix, err := repo.PackObjects(t, seeded.All(), testgit.PackObjectsOptions{
				RevIndex: true,
				Revs:     false,
				Exclude:  nil,
			})
			if err != nil {
				t.Fatalf("PackObjects: %v", err)
			}

			stream, err := os.ReadFile(gitPrefix + ".pack") //nolint:gosec
			if err != nil {
				t.Fatalf("ReadFile pack: %v", err)
			}

			dir := t.TempDir()

			root, err := os.OpenRoot(dir)
			if err != nil {
				t.Fatalf("OpenRoot: %v", err)
			}

			t.Cleanup(func() { _ = root.Close() })

			first, err := ingest.WritePack(root, objectFormat, bytes.NewReader(stream), store.PackWriteOptions{
				ThinBase: nil,
				Progress: nil,
			})
			if err != nil {
				t.Fatalf("first WritePack: %v", err)
			}

			second, err := ingest.WritePack(root, objectFormat, bytes.NewReader(stream), store.PackWriteOptions{
				ThinBase: nil,
				Progress: nil,
			})
			if err != nil {
				t.Fatalf("second WritePack: %v", err)
			}

			if second.PackName != first.PackName {
				t.Fatalf("second PackName = %q, want %q", second.PackName, first.PackName)
			}

			for _, name := range []string{first.PackName, first.IdxName, first.RevName} {
				_, err := os.Stat(filepath.Join(dir, name))
				if err != nil {
					t.Fatalf("missing %q after re-write: %v", name, err)
				}
			}
		})
	}
}

// writePack ingests src into a fresh store directory,
// returning the directory and the ingest result.
func writePack(
	t *testing.T,
	objectFormat id.ObjectFormat,
	src io.Reader,
	opts store.PackWriteOptions,
) (string, ingest.Result) {
	t.Helper()

	dir := t.TempDir()

	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}

	t.Cleanup(func() { _ = root.Close() })

	result, err := ingest.WritePack(root, objectFormat, src, opts)
	if err != nil {
		t.Fatalf("WritePack: %v", err)
	}

	return dir, result
}

// TestWritePackThin verifies that a thin pack is completed from the thin base
// and that git accepts the resulting self-contained pack.
func TestWritePackThin(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, seeded := seedHistory(t, objectFormat)
			thinBase := fullStore(t, repo, objectFormat, seeded)
			stream := thinStream(t, repo, seeded)

			dir, result := writePack(t, objectFormat, bytes.NewReader(stream), store.PackWriteOptions{
				ThinBase: thinBase,
				Progress: nil,
			})

			if !result.ThinFixed {
				t.Fatalf("ThinFixed = false, want true (pack was not thin)")
			}

			_, err := repo.VerifyPack(t, filepath.Join(dir, result.IdxName))
			if err != nil {
				t.Fatalf("VerifyPack on completed pack: %v", err)
			}
		})
	}
}

// TestWritePackThinWithoutBase verifies that a thin pack is rejected
// when no thin base is supplied.
func TestWritePackThinWithoutBase(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, seeded := seedHistory(t, objectFormat)
			stream := thinStream(t, repo, seeded)

			_, err := ingest.WritePack(freshRoot(t), objectFormat, bytes.NewReader(stream), store.PackWriteOptions{
				ThinBase: nil,
				Progress: nil,
			})

			if !errors.Is(err, ingest.ErrThinPackNotPermitted) {
				t.Fatalf("err = %v, want ErrThinPackNotPermitted", err)
			}
		})
	}
}

// TestWritePackThinMissingBase verifies that a thin pack
// whose bases are absent from the thin base
// reports the missing object IDs.
func TestWritePackThinMissingBase(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, seeded := seedHistory(t, objectFormat)
			emptyBase := emptyStore(t, objectFormat)
			stream := thinStream(t, repo, seeded)

			_, err := ingest.WritePack(freshRoot(t), objectFormat, bytes.NewReader(stream), store.PackWriteOptions{
				ThinBase: emptyBase,
				Progress: nil,
			})

			missing, ok := errors.AsType[*ingest.ThinBasesMissingError](err)
			if !ok {
				t.Fatalf("err = %v, want *ThinBasesMissingError", err)
			}

			if len(missing.OIDs) == 0 {
				t.Fatalf("ThinBasesMissingError reported no object IDs")
			}
		})
	}
}

// seedHistory creates one repository with a seeded history.
func seedHistory(t *testing.T, objectFormat id.ObjectFormat) (*testgit.Repo, testgit.Seeded) {
	t.Helper()

	repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
	if err != nil {
		t.Fatalf("NewRepo: %v", err)
	}

	seeded, err := repo.SeedHistory(t)
	if err != nil {
		t.Fatalf("SeedHistory: %v", err)
	}

	return repo, seeded
}

// thinStream produces a thin pack of the tip commit excluding its parent,
// so its deltas reference the omitted parent objects.
func thinStream(t *testing.T, repo *testgit.Repo, seeded testgit.Seeded) []byte {
	t.Helper()

	tip := seeded.Commits[len(seeded.Commits)-1]
	parent := seeded.Commits[len(seeded.Commits)-2]

	stream, err := repo.PackObjectsStdout(t, []id.ObjectID{tip}, testgit.PackObjectsStdoutOptions{
		Revs:    true,
		Thin:    true,
		Exclude: []id.ObjectID{parent},
	})
	if err != nil {
		t.Fatalf("PackObjectsStdout: %v", err)
	}

	return stream
}

// fullStore opens a packed store over a pack of every seeded object.
func fullStore(t *testing.T, repo *testgit.Repo, objectFormat id.ObjectFormat, seeded testgit.Seeded) *packed.Packed {
	t.Helper()

	prefix, err := repo.PackObjects(t, seeded.All(), testgit.PackObjectsOptions{
		RevIndex: false,
		Revs:     false,
		Exclude:  nil,
	})
	if err != nil {
		t.Fatalf("PackObjects: %v", err)
	}

	return openStore(t, filepath.Dir(prefix), objectFormat)
}

// emptyStore opens a packed store over an empty directory.
func emptyStore(t *testing.T, objectFormat id.ObjectFormat) *packed.Packed {
	t.Helper()

	return openStore(t, t.TempDir(), objectFormat)
}

// openStore opens a packed store over dir.
func openStore(t *testing.T, dir string, objectFormat id.ObjectFormat) *packed.Packed {
	t.Helper()

	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}

	t.Cleanup(func() { _ = root.Close() })

	packedStore, err := packed.New(root, objectFormat)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	t.Cleanup(func() { _ = packedStore.Close() })

	return packedStore
}

// freshRoot opens a writable root over a fresh temporary directory.
func freshRoot(t *testing.T) *os.Root {
	t.Helper()

	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}

	t.Cleanup(func() { _ = root.Close() })

	return root
}
