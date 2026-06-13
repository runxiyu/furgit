package tree_test

import (
	"strconv"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/tree"
	"lindenii.org/go/furgit/object/tree/mode"
	"lindenii.org/go/furgit/object/typ"
)

func mixedEntries(tb testing.TB, repo *testgit.Repo) []tree.Entry {
	tb.Helper()

	blobA, err := repo.HashObject(tb, typ.Blob, strings.NewReader("blob-A\n"))
	if err != nil {
		tb.Fatalf("HashObject(blob-A): %v", err)
	}

	blobB, err := repo.HashObject(tb, typ.Blob, strings.NewReader("blob-B\n"))
	if err != nil {
		tb.Fatalf("HashObject(blob-B): %v", err)
	}

	blobC, err := repo.HashObject(tb, typ.Blob, strings.NewReader("blob-C\n"))
	if err != nil {
		tb.Fatalf("HashObject(blob-C): %v", err)
	}

	subTree, err := repo.MkTree(tb, []testgit.TreeEntry{
		{Mode: "100644", Type: typ.Blob, OID: blobA, Name: "leaf"},
	})
	if err != nil {
		tb.Fatalf("MkTree(subtree): %v", err)
	}

	submodule, err := repo.CommitTree(tb, subTree, testgit.CommitTreeOptions{Message: "submodule"})
	if err != nil {
		tb.Fatalf("CommitTree(submodule): %v", err)
	}

	return []tree.Entry{
		{Mode: mode.Regular, Name: []byte("z"), ID: blobA},
		{Mode: mode.Regular, Name: []byte("A"), ID: blobB},
		{Mode: mode.Regular, Name: []byte("aa"), ID: blobC},
		{Mode: mode.Regular, Name: []byte("a0"), ID: blobA},
		{Mode: mode.Regular, Name: []byte("a."), ID: blobC},
		{Mode: mode.Regular, Name: []byte("Z"), ID: blobB},
		{Mode: mode.Regular, Name: []byte("0"), ID: blobA},
		{Mode: mode.Regular, Name: []byte("CAPS"), ID: blobB},
		{Mode: mode.Regular, Name: []byte("caps"), ID: blobC},
		{Mode: mode.Regular, Name: []byte("name with space"), ID: blobB},
		{Mode: mode.Regular, Name: []byte("name.with.dot"), ID: blobA},
		{Mode: mode.Regular, Name: []byte("这是一些非 ASCII 的字符"), ID: blobC},
		{Mode: mode.Regular, Name: []byte("Emoji 👀"), ID: blobC},
		{Mode: mode.Regular, Name: []byte(".hidden"), ID: blobA},
		{Mode: mode.Executable, Name: []byte("exec.sh"), ID: blobB},
		{Mode: mode.Symlink, Name: []byte("sym.link"), ID: blobC},
		{Mode: mode.Gitlink, Name: []byte("submodule"), ID: submodule},
		{Mode: mode.Regular, Name: []byte("dir-"), ID: blobA},
		{Mode: mode.Directory, Name: []byte("dir"), ID: subTree},
		{Mode: mode.Regular, Name: []byte("dir0"), ID: blobB},
	}
}

func mkTreeEntries(entries []tree.Entry) []testgit.TreeEntry {
	out := make([]testgit.TreeEntry, len(entries))
	for i, entry := range entries {
		out[i] = testgit.TreeEntry{
			Mode: strconv.FormatUint(uint64(entry.Mode), 8),
			Type: entry.Mode.ObjectType(),
			OID:  entry.ID,
			Name: string(entry.Name),
		}
	}

	return out
}

func buildTree(tb testing.TB, entries []tree.Entry) *tree.Tree {
	tb.Helper()

	tr := new(tree.Tree)
	for _, entry := range entries {
		err := tr.Insert(entry)
		if err != nil {
			tb.Fatalf("Insert(%q): %v", entry.Name, err)
		}
	}

	return tr
}

func assertGitDecode(tb testing.TB, repo *testgit.Repo, treeID id.ObjectID, got []tree.Entry) {
	tb.Helper()

	want, err := repo.LsTree(tb, treeID)
	if err != nil {
		tb.Fatalf("LsTree: %v", err)
	}

	if len(got) != len(want) {
		tb.Fatalf("entry count = %d, want %d", len(got), len(want))
	}

	for i := range want {
		wantMode, err := strconv.ParseUint(want[i].Mode, 8, 32)
		if err != nil {
			tb.Fatalf("entry[%d] parse git mode %q: %v", i, want[i].Mode, err)
		}

		if uint64(got[i].Mode) != wantMode {
			tb.Fatalf("entry[%d] mode = %o, want %o", i, uint64(got[i].Mode), wantMode)
		}

		if got[i].Mode.ObjectType() != want[i].Type {
			tb.Fatalf("entry[%d] type = %v, want %v", i, got[i].Mode.ObjectType(), want[i].Type)
		}

		if got[i].ID != want[i].OID {
			tb.Fatalf("entry[%d] id = %s, want %s", i, got[i].ID, want[i].OID)
		}

		if string(got[i].Name) != want[i].Name {
			tb.Fatalf("entry[%d] name = %q, want %q", i, got[i].Name, want[i].Name)
		}
	}
}
