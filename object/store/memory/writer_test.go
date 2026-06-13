package memory_test

import (
	"bytes"
	"testing"

	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store/memory"
	"lindenii.org/go/furgit/object/typ"
)

func TestWriteReaderContent(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			store := memory.New(objectFormat)
			content := []byte("memory-content\n")
			raw := append(header.Append(nil, typ.Blob, len(content)), content...)

			gotID, err := store.WriteReaderContent(typ.Blob, len(content), bytes.NewReader(content))
			if err != nil {
				t.Fatalf("WriteReaderContent: %v", err)
			}

			wantID := objectFormat.Sum(raw)
			if gotID != wantID {
				t.Fatalf("WriteReaderContent id = %s, want %s", gotID, wantID)
			}

			gotType, gotContent, err := store.ReadBytesContent(gotID)
			if err != nil {
				t.Fatalf("ReadBytesContent: %v", err)
			}

			if gotType != typ.Blob {
				t.Fatalf("ReadBytesContent type = %v, want %v", gotType, typ.Blob)
			}

			if !bytes.Equal(gotContent, content) {
				t.Fatalf("ReadBytesContent content = %q, want %q", gotContent, content)
			}
		})
	}
}

func TestWriteReaderFull(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			store := memory.New(objectFormat)
			content := []byte("memory-full\n")
			raw := append(header.Append(nil, typ.Blob, len(content)), content...)

			gotID, err := store.WriteReaderFull(bytes.NewReader(raw))
			if err != nil {
				t.Fatalf("WriteReaderFull: %v", err)
			}

			wantID := objectFormat.Sum(raw)
			if gotID != wantID {
				t.Fatalf("WriteReaderFull id = %s, want %s", gotID, wantID)
			}

			gotRaw, err := store.ReadBytesFull(gotID)
			if err != nil {
				t.Fatalf("ReadBytesFull: %v", err)
			}

			if !bytes.Equal(gotRaw, raw) {
				t.Fatalf("ReadBytesFull = %q, want %q", gotRaw, raw)
			}
		})
	}
}

func TestWriteBytes(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			store := memory.New(objectFormat)
			content := []byte("memory-bytes\n")
			raw := append(header.Append(nil, typ.Blob, len(content)), content...)

			gotID, err := store.WriteBytesContent(typ.Blob, content)
			if err != nil {
				t.Fatalf("WriteBytesContent: %v", err)
			}

			wantID := objectFormat.Sum(raw)
			if gotID != wantID {
				t.Fatalf("WriteBytesContent id = %s, want %s", gotID, wantID)
			}

			gotID2, err := store.WriteBytesFull(raw)
			if err != nil {
				t.Fatalf("WriteBytesFull: %v", err)
			}

			if gotID2 != wantID {
				t.Fatalf("WriteBytesFull id = %s, want %s", gotID2, wantID)
			}
		})
	}
}

func TestWriteValidationErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		run  func(store *memory.Memory) error
	}{
		{
			name: "content overflow",
			run: func(store *memory.Memory) error {
				_, err := store.WriteReaderContent(typ.Blob, 1, bytes.NewReader([]byte("hello")))

				return err //nolint:wrapcheck
			},
		},
		{
			name: "content short",
			run: func(store *memory.Memory) error {
				_, err := store.WriteReaderContent(typ.Blob, 5, bytes.NewReader([]byte("x")))

				return err //nolint:wrapcheck
			},
		},
		{
			name: "full malformed header",
			run: func(store *memory.Memory) error {
				_, err := store.WriteReaderFull(bytes.NewReader([]byte("not-a-header")))

				return err //nolint:wrapcheck
			},
		},
		{
			name: "full size mismatch",
			run: func(store *memory.Memory) error {
				_, err := store.WriteReaderFull(bytes.NewReader([]byte("blob 1\x00hello")))

				return err //nolint:wrapcheck
			},
		},
		{
			name: "bytes malformed header",
			run: func(store *memory.Memory) error {
				_, err := store.WriteBytesFull([]byte("not-a-header"))

				return err //nolint:wrapcheck
			},
		},
	}

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					store := memory.New(objectFormat)

					err := tc.run(store)
					if err == nil {
						t.Fatalf("expected error")
					}
				})
			}
		})
	}
}
