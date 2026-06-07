package tag_test

import (
	"bytes"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/tag"
	"lindenii.org/go/furgit/object/typ"
)

func TestParse(t *testing.T) {
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

			treeID, err := repo.MkTree(t, []testgit.MkTreeEntry{
				{Mode: "100644", Type: typ.TypeBlob, OID: blobID, Name: "file.txt"},
			})
			if err != nil {
				t.Fatalf("MkTree: %v", err)
			}

			commitID, err := repo.CommitTree(t, treeID, testgit.CommitTreeOptions{
				Message: "tag target subject\n\nbody",
				Author: testgit.Identity{
					Name:  "Target Author",
					Email: "target-author@example.org",
				},
				Committer: testgit.Identity{
					Name:  "Target Committer",
					Email: "target-committer@example.org",
				},
				AuthorDate:    "1234567890 +0000",
				CommitterDate: "1234567891 +0000",
			})
			if err != nil {
				t.Fatalf("CommitTree: %v", err)
			}

			tagID, err := repo.TagAnnotated(t, "v1.2.3", commitID, testgit.TagAnnotatedOptions{
				Message: "tag subject\n\ntag body",
				Tagger: testgit.Identity{
					Name:  "Test Tagger",
					Email: "tagger@example.org",
				},
				TaggerDate: "1234567999 -0330",
			})
			if err != nil {
				t.Fatalf("TagAnnotated: %v", err)
			}

			rawBody, err := repo.CatFile(t, typ.TypeTag, tagID)
			if err != nil {
				t.Fatalf("CatFile: %v", err)
			}

			parsed, err := tag.Parse(rawBody, objectFormat)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}

			if parsed.TargetID != commitID {
				t.Fatalf("target id = %s, want %s", parsed.TargetID, commitID)
			}

			if parsed.TargetType != typ.TypeCommit {
				t.Fatalf("target type = %v, want %v", parsed.TargetType, typ.TypeCommit)
			}

			if !bytes.Equal(parsed.Name, []byte("v1.2.3")) {
				t.Fatalf("name = %q, want %q", parsed.Name, "v1.2.3")
			}

			if !bytes.Equal(parsed.Tagger.Name, []byte("Test Tagger")) {
				t.Fatalf("tagger name = %q, want %q", parsed.Tagger.Name, "Test Tagger")
			}

			if !bytes.Equal(parsed.Tagger.Email, []byte("tagger@example.org")) {
				t.Fatalf("tagger email = %q, want %q", parsed.Tagger.Email, "tagger@example.org")
			}

			if parsed.Tagger.WhenUnix != 1234567999 {
				t.Fatalf("tagger time = %d, want %d", parsed.Tagger.WhenUnix, int64(1234567999))
			}

			if parsed.Tagger.OffsetMinutes != -210 {
				t.Fatalf("tagger offset = %d, want %d", parsed.Tagger.OffsetMinutes, int32(-210))
			}

			if !bytes.Equal(parsed.Message, []byte("tag subject\n\ntag body\n")) {
				t.Fatalf("message = %q, want %q", parsed.Message, "tag subject\n\ntag body\n")
			}
		})
	}
}

func TestParseEmptyMessageWithoutBlankLine(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			objectID := strings.Repeat("1", objectFormat.HexLen())
			body := "" +
				"object " + objectID + "\n" +
				"type commit\n" +
				"tag empty-message-tag\n" +
				"tagger Test Tagger <tagger@example.org> 1234567890 +0000\n"

			got, err := tag.Parse([]byte(body), objectFormat)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}

			if len(got.Message) != 0 {
				t.Fatalf("message = %q, want empty", got.Message)
			}
		})
	}
}
