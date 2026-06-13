package loose_test

import (
	"bytes"
	"errors"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/store/loose"
	"lindenii.org/go/furgit/object/typ"
)

func TestWrite(t *testing.T) {
	t.Parallel()

	writes := []struct {
		name  string
		write func(looseStore *loose.Loose, content []byte) (id.ObjectID, error)
	}{
		{
			name: "BytesContent",
			write: func(looseStore *loose.Loose, content []byte) (id.ObjectID, error) {
				return looseStore.WriteBytesContent(typ.Blob, content)
			},
		},
		{
			name: "ReaderContent",
			write: func(looseStore *loose.Loose, content []byte) (id.ObjectID, error) {
				return looseStore.WriteReaderContent(typ.Blob, len(content), bytes.NewReader(content))
			},
		},
		{
			name: "BytesFull",
			write: func(looseStore *loose.Loose, content []byte) (id.ObjectID, error) {
				raw := header.Append(nil, typ.Blob, len(content))
				raw = append(raw, content...)

				return looseStore.WriteBytesFull(raw)
			},
		},
		{
			name: "ReaderFull",
			write: func(looseStore *loose.Loose, content []byte) (id.ObjectID, error) {
				raw := header.Append(nil, typ.Blob, len(content))
				raw = append(raw, content...)

				return looseStore.WriteReaderFull(bytes.NewReader(raw))
			},
		},
	}

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			looseStore := openLooseStore(t, repo)

			for _, w := range writes {
				t.Run(w.name, func(t *testing.T) {
					content := []byte("written via " + w.name + "\n")

					want, err := repo.HashObject(t, typ.Blob, bytes.NewReader(content))
					if err != nil {
						t.Fatalf("HashObject: %v", err)
					}

					got, err := w.write(looseStore, content)
					if err != nil {
						t.Fatalf("write: %v", err)
					}

					if got != want {
						t.Fatalf("id = %s, want %s", got, want)
					}

					gotBody, err := repo.CatFile(t, typ.Blob, got)
					if err != nil {
						t.Fatalf("CatFile: %v", err)
					}

					if !bytes.Equal(gotBody, content) {
						t.Fatalf("git cat-file body mismatch")
					}

					regot, err := w.write(looseStore, content)
					if err != nil {
						t.Fatalf("rewrite: %v", err)
					}

					if regot != want {
						t.Fatalf("rewrite id = %s, want %s", regot, want)
					}
				})
			}
		})
	}
}

func TestWriteRejects(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			looseStore := openLooseStore(t, repo)

			t.Run("ContentOverflow", func(t *testing.T) {
				t.Parallel()

				_, err := looseStore.WriteReaderContent(typ.Blob, 1, bytes.NewReader([]byte("hello")))
				if !errors.Is(err, store.ErrInvalidObject) {
					t.Fatalf("err = %v, want ErrInvalidObject", err)
				}
			})

			t.Run("ContentShort", func(t *testing.T) {
				t.Parallel()

				_, err := looseStore.WriteReaderContent(typ.Blob, 5, bytes.NewReader([]byte("x")))
				if !errors.Is(err, store.ErrInvalidObject) {
					t.Fatalf("err = %v, want ErrInvalidObject", err)
				}
			})

			t.Run("FullMalformedHeader", func(t *testing.T) {
				t.Parallel()

				_, err := looseStore.WriteReaderFull(bytes.NewReader([]byte("not-a-header")))
				if !errors.Is(err, store.ErrInvalidObject) {
					t.Fatalf("err = %v, want ErrInvalidObject", err)
				}
			})

			t.Run("FullSizeMismatch", func(t *testing.T) {
				t.Parallel()

				_, err := looseStore.WriteReaderFull(bytes.NewReader([]byte("blob 1\x00hello")))
				if !errors.Is(err, store.ErrInvalidObject) {
					t.Fatalf("err = %v, want ErrInvalidObject", err)
				}
			})
		})
	}
}
