package commit_test

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/commit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/signature"
	"lindenii.org/go/furgit/object/typ"
)

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			blobID, err := repo.HashObject(t, typ.TypeBlob, strings.NewReader("roundtrip\n"))
			if err != nil {
				t.Fatalf("HashObject(blob): %v", err)
			}

			treeID, err := repo.MkTree(t, []testgit.TreeEntry{
				{Mode: "100644", Type: typ.TypeBlob, OID: blobID, Name: "roundtrip.txt"},
			})
			if err != nil {
				t.Fatalf("MkTree: %v", err)
			}

			parent1, err := repo.CommitTree(t, treeID, testgit.CommitTreeOptions{
				Message: "parent one",
				Author: testgit.Identity{
					Name:  "Parent Author",
					Email: "parent-author@example.org",
				},
				Committer: testgit.Identity{
					Name:  "Parent Committer",
					Email: "parent-committer@example.org",
				},
				AuthorDate:    "1234567890 +0000",
				CommitterDate: "1234567890 +0000",
			})
			if err != nil {
				t.Fatalf("CommitTree(parent1): %v", err)
			}

			parent2, err := repo.CommitTree(t, treeID, testgit.CommitTreeOptions{
				Message: "parent two",
				Author: testgit.Identity{
					Name:  "Parent Author",
					Email: "parent-author@example.org",
				},
				Committer: testgit.Identity{
					Name:  "Parent Committer",
					Email: "parent-committer@example.org",
				},
				AuthorDate:    "1234567891 +0000",
				CommitterDate: "1234567891 +0000",
			}, parent1)
			if err != nil {
				t.Fatalf("CommitTree(parent2): %v", err)
			}

			want := &commit.Commit{
				Tree:    treeID,
				Parents: []id.ObjectID{parent1, parent2},
				Author: signature.Signature{
					Name:          []byte("Round Trip Author"),
					Email:         []byte("roundtrip-author@example.org"),
					WhenUnix:      1234567000,
					OffsetMinutes: -210,
				},
				Committer: signature.Signature{
					Name:          []byte("Round Trip Committer"),
					Email:         []byte("roundtrip-committer@example.org"),
					WhenUnix:      1234567999,
					OffsetMinutes: 330,
				},
				Message:  []byte("roundtrip subject\n\nroundtrip body\n\n"),
				ChangeID: "zyxwvutsrqponmlk",
				ExtraHeaders: []commit.ExtraHeader{
					{Key: "encoding", Value: []byte("UTF-8")},
					{Key: "x-test-header", Value: []byte("value")},
				},
			}

			rawBody, err := want.AppendWithoutHeader(nil)
			if err != nil {
				t.Fatalf("AppendWithoutHeader: %v", err)
			}

			roundTripID, err := repo.HashObject(t, typ.TypeCommit, bytes.NewReader(rawBody))
			if err != nil {
				t.Fatalf("HashObject(commit): %v", err)
			}

			err = repo.Fsck(t, testgit.FsckOptions{
				Strict:     true,
				NoDangling: true,
			}, roundTripID)
			if err != nil {
				t.Fatalf("Fsck: %v", err)
			}

			gitBody, err := repo.CatFile(t, typ.TypeCommit, roundTripID)
			if err != nil {
				t.Fatalf("CatFile: %v", err)
			}

			got, err := commit.Parse(gitBody, objectFormat)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}

			assertCommitEqual(t, got, want)
		})
	}
}

func assertCommitEqual(t *testing.T, got *commit.Commit, want *commit.Commit) {
	t.Helper()

	if got.Tree != want.Tree {
		t.Fatalf("tree = %s, want %s", got.Tree, want.Tree)
	}

	if !slices.Equal(got.Parents, want.Parents) {
		t.Fatalf("parents = %v, want %v", got.Parents, want.Parents)
	}

	assertSignatureEqual(t, "author", got.Author, want.Author)
	assertSignatureEqual(t, "committer", got.Committer, want.Committer)

	if !bytes.Equal(got.Message, want.Message) {
		t.Fatalf("message = %q, want %q", got.Message, want.Message)
	}

	if got.ChangeID != want.ChangeID {
		t.Fatalf("change id = %q, want %q", got.ChangeID, want.ChangeID)
	}

	if len(got.ExtraHeaders) != len(want.ExtraHeaders) {
		t.Fatalf("extra header count = %d, want %d", len(got.ExtraHeaders), len(want.ExtraHeaders))
	}

	for i, wantHeader := range want.ExtraHeaders {
		gotHeader := got.ExtraHeaders[i]
		if gotHeader.Key != wantHeader.Key {
			t.Fatalf("extra header[%d] key = %q, want %q", i, gotHeader.Key, wantHeader.Key)
		}

		if !bytes.Equal(gotHeader.Value, wantHeader.Value) {
			t.Fatalf("extra header[%d] value = %q, want %q", i, gotHeader.Value, wantHeader.Value)
		}
	}
}

func assertSignatureEqual(t *testing.T, name string, got signature.Signature, want signature.Signature) {
	t.Helper()

	if !bytes.Equal(got.Name, want.Name) {
		t.Fatalf("%s name = %q, want %q", name, got.Name, want.Name)
	}

	if !bytes.Equal(got.Email, want.Email) {
		t.Fatalf("%s email = %q, want %q", name, got.Email, want.Email)
	}

	if got.WhenUnix != want.WhenUnix {
		t.Fatalf("%s time = %d, want %d", name, got.WhenUnix, want.WhenUnix)
	}

	if got.OffsetMinutes != want.OffsetMinutes {
		t.Fatalf("%s offset = %d, want %d", name, got.OffsetMinutes, want.OffsetMinutes)
	}
}
