package tree_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/tree"
)

func buildGitMktreeInput(entries []tree.TreeEntry) string {
	var b strings.Builder
	for _, e := range entries {
		fmt.Fprintf(&b, "%o %s %s\t%s\n", e.Mode, mktreeTypeFromMode(e.Mode), e.ID.String(), e.Name)
	}

	return b.String()
}

func mktreeTypeFromMode(mode tree.FileMode) string {
	switch mode {
	case tree.FileModeDir:
		return "tree"
	case tree.FileModeRegular, tree.FileModeExecutable, tree.FileModeSymlink:
		return "blob"
	case tree.FileModeGitlink:
		return "commit"
	default:
		return ""
	}
}

func gitLsTreeNames(out []byte) [][]byte {
	if len(out) == 0 {
		return nil
	}

	parts := bytes.Split(out, []byte{0})
	if len(parts) > 0 && len(parts[len(parts)-1]) == 0 {
		parts = parts[:len(parts)-1]
	}

	names := make([][]byte, 0, len(parts))
	for _, name := range parts {
		names = append(names, append([]byte(nil), name...))
	}

	return names
}

func adversarialRootEntries(t *testing.T, testRepo *testgit.TestRepo) []tree.TreeEntry {
	t.Helper()

	blobA := testRepo.HashObject(t, "blob", []byte("blob-A\n"))
	blobB := testRepo.HashObject(t, "blob", []byte("blob-B\n"))
	blobC := testRepo.HashObject(t, "blob", []byte("blob-C\n"))

	subDirA := testRepo.Mktree(t,
		fmt.Sprintf("100644 blob %s\tnested-a.txt\n100755 blob %s\trun-a.sh\n", blobA.String(), blobB.String()))
	subDirB := testRepo.Mktree(t,
		fmt.Sprintf("100644 blob %s\tnested-b.txt\n100644 blob %s\tz-last\n", blobB.String(), blobC.String()))
	subDirC := testRepo.Mktree(t,
		fmt.Sprintf("120000 blob %s\tlink-c\n100644 blob %s\tchild\n", blobC.String(), blobA.String()))
	subDirD := testRepo.Mktree(t,
		fmt.Sprintf("100644 blob %s\tleaf\n", blobA.String()))

	return []tree.TreeEntry{
		{Mode: tree.FileModeRegular, Name: []byte("z"), ID: blobA},
		{Mode: tree.FileModeRegular, Name: []byte("A"), ID: blobB},
		{Mode: tree.FileModeRegular, Name: []byte("aa"), ID: blobC},
		{Mode: tree.FileModeRegular, Name: []byte("a0"), ID: blobA},
		{Mode: tree.FileModeRegular, Name: []byte("a-"), ID: blobB},
		{Mode: tree.FileModeRegular, Name: []byte("a."), ID: blobC},
		{Mode: tree.FileModeRegular, Name: []byte("a_"), ID: blobA},
		{Mode: tree.FileModeRegular, Name: []byte("a~"), ID: blobB},
		{Mode: tree.FileModeRegular, Name: []byte("Z"), ID: blobC},
		{Mode: tree.FileModeRegular, Name: []byte("0"), ID: blobA},
		{Mode: tree.FileModeRegular, Name: []byte("9"), ID: blobB},
		{Mode: tree.FileModeRegular, Name: []byte("00"), ID: blobC},
		{Mode: tree.FileModeRegular, Name: []byte("这是一些非 ASCII 的字符"), ID: blobC},
		{Mode: tree.FileModeRegular, Name: []byte("𲰼是新进入 Unicode 的字符"), ID: blobC},
		{Mode: tree.FileModeRegular, Name: []byte("Emoji 👀"), ID: blobC},
		{Mode: tree.FileModeRegular, Name: []byte("_"), ID: blobA},
		{Mode: tree.FileModeRegular, Name: []byte("-dash"), ID: blobB},
		{Mode: tree.FileModeRegular, Name: []byte("dot.file"), ID: blobC},
		{Mode: tree.FileModeRegular, Name: []byte(".hidden"), ID: blobA},
		{Mode: tree.FileModeRegular, Name: []byte("CAPS"), ID: blobB},
		{Mode: tree.FileModeRegular, Name: []byte("caps"), ID: blobC},
		{Mode: tree.FileModeRegular, Name: []byte("mixCase"), ID: blobA},
		{Mode: tree.FileModeRegular, Name: []byte("name with space"), ID: blobB},
		{Mode: tree.FileModeRegular, Name: []byte("name-with-dash"), ID: blobC},
		{Mode: tree.FileModeRegular, Name: []byte("name.with.dot"), ID: blobA},
		{Mode: tree.FileModeRegular, Name: []byte("name_with_underscore"), ID: blobB},
		{Mode: tree.FileModeRegular, Name: []byte("tilde~name"), ID: blobC},
		{Mode: tree.FileModeRegular, Name: []byte("brace{name}"), ID: blobA},
		{Mode: tree.FileModeRegular, Name: []byte("plus+name"), ID: blobB},
		{Mode: tree.FileModeRegular, Name: []byte("equal=name"), ID: blobC},
		{Mode: tree.FileModeRegular, Name: []byte("at@name"), ID: blobA},
		{Mode: tree.FileModeRegular, Name: []byte("percent%name"), ID: blobB},
		{Mode: tree.FileModeRegular, Name: []byte("caret^name"), ID: blobC},
		{Mode: tree.FileModeRegular, Name: []byte("comma,name"), ID: blobA},
		{Mode: tree.FileModeRegular, Name: []byte("semi;name"), ID: blobB},
		{Mode: tree.FileModeRegular, Name: []byte("paren(name)"), ID: blobC},
		{Mode: tree.FileModeRegular, Name: []byte("bracket[name]"), ID: blobA},
		{Mode: tree.FileModeExecutable, Name: []byte("exec.sh"), ID: blobB},
		{Mode: tree.FileModeSymlink, Name: []byte("sym.link"), ID: blobC},
		{Mode: tree.FileModeDir, Name: []byte("dir"), ID: subDirA},
		{Mode: tree.FileModeDir, Name: []byte("dir0"), ID: subDirB},
		{Mode: tree.FileModeDir, Name: []byte("dir.space"), ID: subDirC},
		{Mode: tree.FileModeDir, Name: []byte("x"), ID: subDirD},
	}
}
