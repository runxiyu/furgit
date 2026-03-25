package resolve_test

import (
	"errors"
	"io/fs"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/resolve"
	"codeberg.org/lindenii/furgit/repository"
)

func TestTreeFS(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		t.Parallel()

		repoData := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo})
		repoData.WriteFile(t, "plain.txt", []byte("plain\n"), 0o644)
		repoData.WriteFileAll(t, "dir/exec.sh", []byte("#!/bin/sh\nexit 0\n"), 0o755, 0o755)
		repoData.SymbolicRef(t, "HEAD", "refs/heads/main")
		_ = repoData.Run(t, "add", ".")
		treeHex := repoData.Run(t, "write-tree")

		treeID, err := objectid.ParseHex(algo, treeHex)
		if err != nil {
			t.Fatalf("ParseHex(write-tree): %v", err)
		}

		commitID := repoData.CommitTree(t, treeID, "treefs")

		root := repoData.OpenGitRoot(t)

		repo, err := repository.Open(root)
		if err != nil {
			t.Fatalf("repository.Open: %v", err)
		}

		defer func() { _ = repo.Close() }()

		resolver := resolve.New(repo.Objects())

		treeFS, err := resolver.TreeFS(commitID)
		if err != nil {
			t.Fatalf("resolver.TreeFS: %v", err)
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

		entry, ok := info.Sys().(object.TreeEntry)
		if !ok {
			t.Fatalf("Stat(plain.txt).Sys() type = %T, want object.TreeEntry", info.Sys())
		}

		if entry.Mode != object.FileModeRegular {
			t.Fatalf("Stat(plain.txt).Sys().Mode = %o, want %o", entry.Mode, object.FileModeRegular)
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

		var pathErr *fs.PathError
		if !errors.As(err, &pathErr) {
			t.Fatalf("ReadFile(dir) err type = %T, want *fs.PathError", err)
		}
	})
}
