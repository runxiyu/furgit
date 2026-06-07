package commit_test

import (
	"bytes"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/commit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
)

func TestParse(t *testing.T) {
	t.Parallel()

	commitTreeOptions := func(message string) testgit.CommitTreeOptions {
		return testgit.CommitTreeOptions{
			Message: message,
			Author: testgit.Identity{
				Name:  "Test Author",
				Email: "author@example.org",
			},
			Committer: testgit.Identity{
				Name:  "Test Committer",
				Email: "committer@example.org",
			},
			AuthorDate:    "1234567890 -0730",
			CommitterDate: "1234567999 +0545",
		}
	}

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			blobID, err := repo.HashObject(t, typ.TypeBlob, strings.NewReader("content\n"))
			if err != nil {
				t.Fatalf("HashObject: %v", err)
			}

			treeID, err := repo.MkTree(t, []testgit.MkTreeEntry{
				{Mode: "100644", Type: typ.TypeBlob, OID: blobID, Name: "file.txt"},
			})
			if err != nil {
				t.Fatalf("MkTree: %v", err)
			}

			rootID, err := repo.CommitTree(t, treeID, commitTreeOptions("root subject\n\nroot body"))
			if err != nil {
				t.Fatalf("CommitTree(root): %v", err)
			}

			childID, err := repo.CommitTree(t, treeID, commitTreeOptions("child subject\n\nchild body"), rootID)
			if err != nil {
				t.Fatalf("CommitTree(child): %v", err)
			}

			sideID, err := repo.CommitTree(t, treeID, commitTreeOptions("side subject\n\nside body"))
			if err != nil {
				t.Fatalf("CommitTree(side): %v", err)
			}

			mergeID, err := repo.CommitTree(t, treeID, commitTreeOptions("merge subject\n\nmerge body"), childID, sideID)
			if err != nil {
				t.Fatalf("CommitTree(merge): %v", err)
			}

			for _, tc := range []struct {
				name    string
				oid     id.ObjectID
				parents []id.ObjectID
				message []byte
			}{
				{
					name:    "root",
					oid:     rootID,
					message: []byte("root subject\n\nroot body\n"),
				},
				{
					name:    "child",
					oid:     childID,
					parents: []id.ObjectID{rootID},
					message: []byte("child subject\n\nchild body\n"),
				},
				{
					name:    "merge",
					oid:     mergeID,
					parents: []id.ObjectID{childID, sideID},
					message: []byte("merge subject\n\nmerge body\n"),
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					rawBody, err := repo.CatFile(t, typ.TypeCommit, tc.oid)
					if err != nil {
						t.Fatalf("CatFile: %v", err)
					}

					parsed, err := commit.Parse(rawBody, objectFormat)
					if err != nil {
						t.Fatalf("Parse: %v", err)
					}

					if parsed.Tree != treeID {
						t.Fatalf("tree id mismatch: got %s want %s", parsed.Tree, treeID)
					}

					if len(parsed.Parents) != len(tc.parents) {
						t.Fatalf("parent count = %d, want %d", len(parsed.Parents), len(tc.parents))
					}

					for i, parent := range tc.parents {
						if parsed.Parents[i] != parent {
							t.Fatalf("parent[%d] = %s, want %s", i, parsed.Parents[i], parent)
						}
					}

					if !bytes.Equal(parsed.Author.Name, []byte("Test Author")) {
						t.Fatalf("author name = %q, want %q", parsed.Author.Name, "Test Author")
					}

					if !bytes.Equal(parsed.Author.Email, []byte("author@example.org")) {
						t.Fatalf("author email = %q, want %q", parsed.Author.Email, "author@example.org")
					}

					if parsed.Author.WhenUnix != 1234567890 {
						t.Fatalf("author time = %d, want %d", parsed.Author.WhenUnix, int64(1234567890))
					}

					if parsed.Author.OffsetMinutes != -450 {
						t.Fatalf("author offset = %d, want %d", parsed.Author.OffsetMinutes, int32(-450))
					}

					if !bytes.Equal(parsed.Committer.Name, []byte("Test Committer")) {
						t.Fatalf("committer name = %q, want %q", parsed.Committer.Name, "Test Committer")
					}

					if !bytes.Equal(parsed.Committer.Email, []byte("committer@example.org")) {
						t.Fatalf("committer email = %q, want %q", parsed.Committer.Email, "committer@example.org")
					}

					if parsed.Committer.WhenUnix != 1234567999 {
						t.Fatalf("committer time = %d, want %d", parsed.Committer.WhenUnix, int64(1234567999))
					}

					if parsed.Committer.OffsetMinutes != 345 {
						t.Fatalf("committer offset = %d, want %d", parsed.Committer.OffsetMinutes, int32(345))
					}

					if !bytes.Equal(parsed.Message, tc.message) {
						t.Fatalf("message = %q, want %q", parsed.Message, tc.message)
					}
				})
			}
		})
	}
}
