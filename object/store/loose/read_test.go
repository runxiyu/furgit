package loose_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/typ"
)

func TestRead(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			objects := gitOracleObjects(t, repo)
			looseStore := openLooseStore(t, repo)

			t.Run("BytesFull", func(t *testing.T) {
				t.Parallel()

				for _, o := range objects {
					got, err := looseStore.ReadBytesFull(o.id)
					if err != nil {
						t.Fatalf("%s: ReadBytesFull: %v", o.name, err)
					}

					if !bytes.Equal(got, o.raw) {
						t.Fatalf("%s: ReadBytesFull mismatch", o.name)
					}
				}
			})

			t.Run("BytesContent", func(t *testing.T) {
				t.Parallel()

				for _, o := range objects {
					gotType, gotBody, err := looseStore.ReadBytesContent(o.id)
					if err != nil {
						t.Fatalf("%s: ReadBytesContent: %v", o.name, err)
					}

					if gotType != o.ty {
						t.Fatalf("%s: ReadBytesContent type = %v, want %v", o.name, gotType, o.ty)
					}

					if !bytes.Equal(gotBody, o.body) {
						t.Fatalf("%s: ReadBytesContent body mismatch", o.name)
					}
				}
			})

			t.Run("Header", func(t *testing.T) {
				t.Parallel()

				for _, o := range objects {
					gotType, gotSize, err := looseStore.ReadHeader(o.id)
					if err != nil {
						t.Fatalf("%s: ReadHeader: %v", o.name, err)
					}

					if gotType != o.ty {
						t.Fatalf("%s: ReadHeader type = %v, want %v", o.name, gotType, o.ty)
					}

					if gotSize != uint64(len(o.body)) {
						t.Fatalf("%s: ReadHeader size = %d, want %d", o.name, gotSize, len(o.body))
					}
				}
			})

			t.Run("ReaderFull", func(t *testing.T) {
				t.Parallel()

				for _, o := range objects {
					reader, err := looseStore.ReadReaderFull(o.id)
					if err != nil {
						t.Fatalf("%s: ReadReaderFull: %v", o.name, err)
					}

					got, err := io.ReadAll(reader)
					if err != nil {
						_ = reader.Close()

						t.Fatalf("%s: ReadReaderFull ReadAll: %v", o.name, err)
					}

					err = reader.Close()
					if err != nil {
						t.Fatalf("%s: ReadReaderFull Close: %v", o.name, err)
					}

					if !bytes.Equal(got, o.raw) {
						t.Fatalf("%s: ReadReaderFull mismatch", o.name)
					}
				}
			})

			t.Run("ReaderContent", func(t *testing.T) {
				t.Parallel()

				for _, o := range objects {
					gotType, gotSize, reader, err := looseStore.ReadReaderContent(o.id)
					if err != nil {
						t.Fatalf("%s: ReadReaderContent: %v", o.name, err)
					}

					got, err := io.ReadAll(reader)
					if err != nil {
						_ = reader.Close()

						t.Fatalf("%s: ReadReaderContent ReadAll: %v", o.name, err)
					}

					err = reader.Close()
					if err != nil {
						t.Fatalf("%s: ReadReaderContent Close: %v", o.name, err)
					}

					if gotType != o.ty {
						t.Fatalf("%s: ReadReaderContent type = %v, want %v", o.name, gotType, o.ty)
					}

					if gotSize != uint64(len(o.body)) {
						t.Fatalf("%s: ReadReaderContent size = %d, want %d", o.name, gotSize, len(o.body))
					}

					if !bytes.Equal(got, o.body) {
						t.Fatalf("%s: ReadReaderContent mismatch", o.name)
					}
				}
			})
		})
	}
}

func TestReadNotFound(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			looseStore := openLooseStore(t, repo)

			missingID, err := objectFormat.FromString(strings.Repeat("0", objectFormat.HexLen()))
			if err != nil {
				t.Fatalf("FromString(missing): %v", err)
			}

			_, err = looseStore.ReadBytesFull(missingID)
			if !errors.Is(err, store.ErrObjectNotFound) {
				t.Fatalf("ReadBytesFull not-found = %v", err)
			}

			_, _, err = looseStore.ReadBytesContent(missingID)
			if !errors.Is(err, store.ErrObjectNotFound) {
				t.Fatalf("ReadBytesContent not-found = %v", err)
			}

			_, _, err = looseStore.ReadHeader(missingID)
			if !errors.Is(err, store.ErrObjectNotFound) {
				t.Fatalf("ReadHeader not-found = %v", err)
			}

			_, err = looseStore.ReadReaderFull(missingID)
			if !errors.Is(err, store.ErrObjectNotFound) {
				t.Fatalf("ReadReaderFull not-found = %v", err)
			}

			_, _, _, err = looseStore.ReadReaderContent(missingID)
			if !errors.Is(err, store.ErrObjectNotFound) {
				t.Fatalf("ReadReaderContent not-found = %v", err)
			}

			otherFormat := objectFormat

			for _, candidate := range id.SupportedObjectFormats() {
				if candidate != objectFormat {
					otherFormat = candidate

					break
				}
			}

			if otherFormat == objectFormat {
				return
			}

			mismatchID, err := otherFormat.FromString(strings.Repeat("1", otherFormat.HexLen()))
			if err != nil {
				t.Fatalf("FromString(mismatch): %v", err)
			}

			_, err = looseStore.ReadBytesFull(mismatchID)
			if !errors.Is(err, id.ErrInvalidObjectFormat) {
				t.Fatalf("ReadBytesFull format mismatch = %v, want ErrInvalidObjectFormat", err)
			}
		})
	}
}

func TestReadCorruptTrailer(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			looseStore := openLooseStore(t, repo)

			content := []byte("corrupt-trailer-check\n")

			objectID, err := looseStore.WriteBytesContent(typ.Blob, content)
			if err != nil {
				t.Fatalf("WriteBytesContent: %v", err)
			}

			corruptLooseObjectTrailer(t, repo, objectID)

			// Stops before the trailer.
			ty, size, err := looseStore.ReadHeader(objectID)
			if err != nil {
				t.Fatalf("ReadHeader: %v", err)
			}

			if ty != typ.Blob {
				t.Fatalf("ReadHeader type = %v, want %v", ty, typ.Blob)
			}

			if size != uint64(len(content)) {
				t.Fatalf("ReadHeader size = %d, want %d", size, len(content))
			}

			// Consumes the whole stream.
			_, err = looseStore.ReadBytesFull(objectID)
			if err == nil {
				t.Fatalf("ReadBytesFull on corrupt trailer succeeded")
			}
		})
	}
}
