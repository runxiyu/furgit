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

func TestAppendGitFsck(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			blobID, err := repo.HashObject(t, typ.TypeBlob, strings.NewReader("content\n"))
			if err != nil {
				t.Fatalf("HashObject(blob): %v", err)
			}

			tagObject := &tag.Tag{
				TargetID:   blobID,
				TargetType: typ.TypeBlob,
				Name:       []byte("blob-tag"),
				Tagger: signature.Signature{
					Name:          []byte("Test Tagger"),
					Email:         []byte("tagger@example.org"),
					WhenUnix:      1234567890,
					OffsetMinutes: 0,
				},
				Message: []byte("subject\n\nbody\n"),
			}

			rawBody, err := tagObject.AppendWithoutHeader(nil)
			if err != nil {
				t.Fatalf("AppendWithoutHeader: %v", err)
			}

			tagID, err := repo.HashObject(t, typ.TypeTag, bytes.NewReader(rawBody))
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

			gitBody, err := repo.CatFile(t, typ.TypeTag, tagID)
			if err != nil {
				t.Fatalf("CatFile(tag): %v", err)
			}

			parsed, err := tag.Parse(gitBody, objectFormat)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}

			assertTagEqual(t, parsed, tagObject)
		})
	}
}
