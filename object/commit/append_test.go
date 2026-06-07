package commit_test

import (
	"bytes"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/commit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/signature"
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

			treeID, err := repo.MkTree(t, []testgit.MkTreeEntry{
				{Mode: "100644", Type: typ.TypeBlob, OID: blobID, Name: "file.txt"},
			})
			if err != nil {
				t.Fatalf("MkTree: %v", err)
			}

			commitObject := &commit.Commit{
				Tree: treeID,
				Author: signature.Signature{
					Name:          []byte("Test Author"),
					Email:         []byte("author@example.org"),
					WhenUnix:      1234567890,
					OffsetMinutes: 0,
				},
				Committer: signature.Signature{
					Name:          []byte("Test Committer"),
					Email:         []byte("committer@example.org"),
					WhenUnix:      1234567890,
					OffsetMinutes: 0,
				},
				Message: []byte("subject\n\nbody\n"),
			}

			rawBody, err := commitObject.AppendWithoutHeader(nil)
			if err != nil {
				t.Fatalf("AppendWithoutHeader: %v", err)
			}

			commitID, err := repo.HashObject(t, typ.TypeCommit, bytes.NewReader(rawBody))
			if err != nil {
				t.Fatalf("HashObject(commit): %v", err)
			}

			err = repo.Fsck(t, testgit.FsckOptions{
				Strict:     true,
				NoDangling: true,
			}, commitID)
			if err != nil {
				t.Fatalf("Fsck: %v", err)
			}

			gitTree, err := repo.ShowFormat(t, commitID, "%T")
			if err != nil {
				t.Fatalf("ShowFormat(%%T): %v", err)
			}

			if got, want := strings.TrimSuffix(string(gitTree), "\n"), treeID.String(); got != want {
				t.Fatalf("tree = %q, want %q", got, want)
			}

			gitParents, err := repo.ShowFormat(t, commitID, "%P")
			if err != nil {
				t.Fatalf("ShowFormat(%%P): %v", err)
			}

			if got := strings.TrimSuffix(string(gitParents), "\n"); got != "" {
				t.Fatalf("parents = %q, want none", got)
			}

			gitAuthorName, err := repo.ShowFormat(t, commitID, "%an")
			if err != nil {
				t.Fatalf("ShowFormat(%%an): %v", err)
			}

			if got, want := strings.TrimSuffix(string(gitAuthorName), "\n"), "Test Author"; got != want {
				t.Fatalf("author name = %q, want %q", got, want)
			}

			gitAuthorEmail, err := repo.ShowFormat(t, commitID, "%ae")
			if err != nil {
				t.Fatalf("ShowFormat(%%ae): %v", err)
			}

			if got, want := strings.TrimSuffix(string(gitAuthorEmail), "\n"), "author@example.org"; got != want {
				t.Fatalf("author email = %q, want %q", got, want)
			}

			gitAuthorTime, err := repo.ShowFormat(t, commitID, "%at")
			if err != nil {
				t.Fatalf("ShowFormat(%%at): %v", err)
			}

			if got, want := strings.TrimSuffix(string(gitAuthorTime), "\n"), "1234567890"; got != want {
				t.Fatalf("author time = %q, want %q", got, want)
			}

			gitCommitterName, err := repo.ShowFormat(t, commitID, "%cn")
			if err != nil {
				t.Fatalf("ShowFormat(%%cn): %v", err)
			}

			if got, want := strings.TrimSuffix(string(gitCommitterName), "\n"), "Test Committer"; got != want {
				t.Fatalf("committer name = %q, want %q", got, want)
			}

			gitCommitterEmail, err := repo.ShowFormat(t, commitID, "%ce")
			if err != nil {
				t.Fatalf("ShowFormat(%%ce): %v", err)
			}

			if got, want := strings.TrimSuffix(string(gitCommitterEmail), "\n"), "committer@example.org"; got != want {
				t.Fatalf("committer email = %q, want %q", got, want)
			}

			gitCommitterTime, err := repo.ShowFormat(t, commitID, "%ct")
			if err != nil {
				t.Fatalf("ShowFormat(%%ct): %v", err)
			}

			if got, want := strings.TrimSuffix(string(gitCommitterTime), "\n"), "1234567890"; got != want {
				t.Fatalf("committer time = %q, want %q", got, want)
			}

			gitSubject, err := repo.ShowFormat(t, commitID, "%s")
			if err != nil {
				t.Fatalf("ShowFormat(%%s): %v", err)
			}

			if got, want := strings.TrimSuffix(string(gitSubject), "\n"), "subject"; got != want {
				t.Fatalf("subject = %q, want %q", got, want)
			}

			gitBody, err := repo.ShowFormat(t, commitID, "%b")
			if err != nil {
				t.Fatalf("ShowFormat(%%b): %v", err)
			}

			if got, want := bytes.TrimSuffix(gitBody, []byte{'\n'}), []byte("body\n"); !bytes.Equal(got, want) {
				t.Fatalf("body = %q, want %q", got, want)
			}
		})
	}
}
