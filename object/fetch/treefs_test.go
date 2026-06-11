package fetch_test

import (
	"errors"
	"io/fs"
	"testing"

	"lindenii.org/go/furgit/object/commit"
	"lindenii.org/go/furgit/object/fetch"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/signature"
	"lindenii.org/go/furgit/object/store/memory"
	"lindenii.org/go/furgit/object/tree"
	"lindenii.org/go/furgit/object/tree/mode"
	"lindenii.org/go/furgit/object/typ"
)

func TestTreeFS(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			store := memory.New(objectFormat)

			plainID, err := store.WriteBytesContent(typ.Blob, []byte("plain\n"))
			if err != nil {
				t.Fatalf("WriteBytesContent(plain.txt): %v", err)
			}

			execID, err := store.WriteBytesContent(typ.Blob, []byte("#!/bin/sh\nexit 0\n"))
			if err != nil {
				t.Fatalf("WriteBytesContent(exec.sh): %v", err)
			}

			subTreeID := writeTree(t, store, []tree.Entry{
				{Mode: mode.Executable, Name: "exec.sh", ID: execID},
			})

			rootTreeID := writeTree(t, store, []tree.Entry{
				{Mode: mode.Regular, Name: "plain.txt", ID: plainID},
				{Mode: mode.Directory, Name: "dir", ID: subTreeID},
			})

			commitID := writeCommit(t, store, rootTreeID)

			fetcher := fetch.New(store)

			treeFS, err := fetcher.TreeFS(commitID)
			if err != nil {
				t.Fatalf("fetcher.TreeFS: %v", err)
			}

			content, err := treeFS.ReadFile("plain.txt")
			if err != nil {
				t.Fatalf("ReadFile(plain.txt): %v", err)
			}

			if string(content) != "plain\n" {
				t.Fatalf("ReadFile(plain.txt) = %q, want %q", string(content), "plain\n")
			}

			entries, err := treeFS.ReadDir(".")
			if err != nil {
				t.Fatalf("ReadDir(.): %v", err)
			}

			if len(entries) != 2 {
				t.Fatalf("len(ReadDir(.)) = %d, want 2", len(entries))
			}

			info, err := treeFS.Stat("plain.txt")
			if err != nil {
				t.Fatalf("Stat(plain.txt): %v", err)
			}

			entry, ok := info.Sys().(tree.Entry)
			if !ok {
				t.Fatalf("Stat(plain.txt).Sys() type = %T, want tree.Entry", info.Sys())
			}

			if entry.Mode != mode.Regular {
				t.Fatalf("Stat(plain.txt).Sys().Mode = %o, want %o", entry.Mode, mode.Regular)
			}

			subFS, err := treeFS.Sub("dir")
			if err != nil {
				t.Fatalf("Sub(dir): %v", err)
			}

			subReadFileFS, ok := subFS.(fs.ReadFileFS)
			if !ok {
				t.Fatalf("Sub(dir) type does not implement fs.ReadFileFS")
			}

			subContent, err := subReadFileFS.ReadFile("exec.sh")
			if err != nil {
				t.Fatalf("Sub(dir).ReadFile(exec.sh): %v", err)
			}

			if string(subContent) != "#!/bin/sh\nexit 0\n" {
				t.Fatalf("Sub(dir).ReadFile(exec.sh) = %q", string(subContent))
			}

			_, err = treeFS.ReadFile("dir")
			if err == nil {
				t.Fatal("ReadFile(dir) unexpectedly succeeded")
			}

			if _, ok := errors.AsType[*fs.PathError](err); !ok {
				t.Fatalf("ReadFile(dir) err type = %T, want *fs.PathError", err)
			}
		})
	}
}

// writeTree builds a tree object from entries, writes it to store,
// and returns its object ID.
func writeTree(t *testing.T, store *memory.Memory, entries []tree.Entry) id.ObjectID {
	t.Helper()

	tr := new(tree.Tree)
	for _, entry := range entries {
		err := tr.Insert(entry)
		if err != nil {
			t.Fatalf("tree.Insert(%q): %v", entry.Name, err)
		}
	}

	body, err := tr.AppendWithoutHeader(nil)
	if err != nil {
		t.Fatalf("tree.AppendWithoutHeader: %v", err)
	}

	treeID, err := store.WriteBytesContent(typ.Tree, body)
	if err != nil {
		t.Fatalf("WriteBytesContent(tree): %v", err)
	}

	return treeID
}

// writeCommit builds a commit object pointing at tree, writes it to store,
// and returns its object ID.
func writeCommit(t *testing.T, store *memory.Memory, tree id.ObjectID) id.ObjectID {
	t.Helper()

	who := signature.Signature{
		Name:          []byte("Test Author"),
		Email:         []byte("author@example.org"),
		WhenUnix:      1234567890,
		OffsetMinutes: 0,
	}

	body, err := (&commit.Commit{
		Tree:      tree,
		Author:    who,
		Committer: who,
		Message:   []byte("treefs\n"),
	}).AppendWithoutHeader(nil)
	if err != nil {
		t.Fatalf("commit.AppendWithoutHeader: %v", err)
	}

	commitID, err := store.WriteBytesContent(typ.Commit, body)
	if err != nil {
		t.Fatalf("WriteBytesContent(commit): %v", err)
	}

	return commitID
}
