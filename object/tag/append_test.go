package tag_test

import (
	"bytes"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/signature"
	"lindenii.org/go/furgit/object/tag"
	"lindenii.org/go/furgit/object/typ"
)

func TestAppend(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			blobID, err := repo.HashObject(t, typ.Blob, strings.NewReader("content\n"))
			if err != nil {
				t.Fatalf("HashObject(blob): %v", err)
			}

			tagObject := &tag.Tag{
				TargetID:   blobID,
				TargetType: typ.Blob,
				Name:       []byte("blob-tag"),
				Tagger: signature.Signature{
					Name:          []byte("Test Tagger"),
					Email:         []byte("tagger@example.org"),
					WhenUnix:      1234567890,
					OffsetMinutes: -210,
				},
				Message: []byte("subject\n\nbody\n"),
			}

			rawBody, err := tagObject.AppendWithoutHeader(nil)
			if err != nil {
				t.Fatalf("AppendWithoutHeader: %v", err)
			}

			tagID, err := repo.HashObject(t, typ.Tag, bytes.NewReader(rawBody))
			if err != nil {
				t.Fatalf("HashObject(tag): %v", err)
			}

			err = repo.Fsck(t, testgit.FsckOptions{
				Strict:     true,
				NoDangling: true,
			}, tagID)
			if err != nil {
				t.Fatalf("Fsck: %v", err)
			}

			const ref = "refs/tags/oracle"

			err = repo.UpdateRef(t, ref, tagID)
			if err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			for _, field := range []struct {
				name   string
				format string
				want   string
			}{
				{name: "object", format: "%(object)", want: blobID.String()},
				{name: "type", format: "%(type)", want: "blob"},
				{name: "tag", format: "%(tag)", want: "blob-tag"},
				{name: "taggername", format: "%(taggername)", want: "Test Tagger"},
				{name: "taggeremail", format: "%(taggeremail:trim)", want: "tagger@example.org"},
				{name: "taggerdate", format: "%(taggerdate:raw)", want: "1234567890 -0330"},
				{name: "subject", format: "%(contents:subject)", want: "subject"},
				{name: "body", format: "%(contents:body)", want: "body\n"},
			} {
				got, err := repo.ForEachRefFormat(t, ref, field.format)
				if err != nil {
					t.Fatalf("ForEachRefFormat(%s): %v", field.format, err)
				}

				if trimmed := strings.TrimSuffix(string(got), "\n"); trimmed != field.want {
					t.Fatalf("%s = %q, want %q", field.name, trimmed, field.want)
				}
			}
		})
	}
}
