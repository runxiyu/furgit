package files_test

import (
	"strings"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref/store"
)

func TestPackedRewriteTraitPropagation(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			aID := objectFormat.Sum([]byte("a"))
			bID := objectFormat.Sum([]byte("b"))
			zID := objectFormat.Sum([]byte("z"))
			peeledID := objectFormat.Sum([]byte("peeled"))

			t.Run("no traits stay absent and output is sorted", func(t *testing.T) {
				t.Parallel()

				repo := newFilesRepo(t, objectFormat)
				gitdir := openGitDirRoot(t, repo)

				content := zID.String() + " refs/heads/z\n" +
					bID.String() + " refs/heads/b\n" +
					aID.String() + " refs/heads/a\n"

				err := gitdir.WriteFile("packed-refs", []byte(content), 0o644)
				if err != nil {
					t.Fatalf("WriteFile: %v", err)
				}

				filesStore := openStore(t, repo, objectFormat)

				commitOps(t, filesStore, func(tx store.Transaction) {
					err := tx.Delete("refs/heads/z", zID)
					if err != nil {
						t.Fatalf("Delete: %v", err)
					}
				})

				rewritten, err := gitdir.ReadFile("packed-refs")
				if err != nil {
					t.Fatalf("ReadFile: %v", err)
				}

				lines := strings.Split(strings.TrimSuffix(string(rewritten), "\n"), "\n")

				expected := []string{
					"# pack-refs with: sorted",
					aID.String() + " refs/heads/a",
					bID.String() + " refs/heads/b",
				}

				if len(lines) != len(expected) {
					t.Fatalf("rewritten = %q, want %q", lines, expected)
				}

				for i := range expected {
					if lines[i] != expected[i] {
						t.Fatalf("line %d = %q, want %q", i, lines[i], expected[i])
					}
				}
			})

			t.Run("full traits preserved with peel lines", func(t *testing.T) {
				t.Parallel()

				repo := newFilesRepo(t, objectFormat)
				gitdir := openGitDirRoot(t, repo)

				content := "# pack-refs with: peeled fully-peeled sorted \n" +
					aID.String() + " refs/heads/a\n" +
					bID.String() + " refs/tags/v1\n" +
					"^" + peeledID.String() + "\n"

				err := gitdir.WriteFile("packed-refs", []byte(content), 0o644)
				if err != nil {
					t.Fatalf("WriteFile: %v", err)
				}

				filesStore := openStore(t, repo, objectFormat)

				commitOps(t, filesStore, func(tx store.Transaction) {
					err := tx.Delete("refs/heads/a", aID)
					if err != nil {
						t.Fatalf("Delete: %v", err)
					}
				})

				rewritten, err := gitdir.ReadFile("packed-refs")
				if err != nil {
					t.Fatalf("ReadFile: %v", err)
				}

				expected := "# pack-refs with: peeled fully-peeled sorted\n" +
					bID.String() + " refs/tags/v1\n" +
					"^" + peeledID.String() + "\n"

				if string(rewritten) != expected {
					t.Fatalf("rewritten = %q, want %q", rewritten, expected)
				}
			})
		})
	}
}
