package furgit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCurrentRepoDepthFirstEnumeration(t *testing.T) {
	gitDir := filepath.Join(".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Skip("no .git directory found in current repo")
	}

	repo, err := OpenRepository(gitDir)
	if err != nil {
		t.Skipf("failed to open current .git directory: %v", err)
	}
	defer func() { _ = repo.Close() }()

	headHash, err := repo.ResolveRefFully("HEAD")
	if err != nil {
		t.Fatalf("failed to resolve HEAD: %v", err)
	}

	visited := make(map[Hash]bool)
	var queue []Hash
	queue = append(queue, headHash)

	objectsRead := 0

	for len(queue) > 0 {
		hash := queue[0]
		queue = queue[1:]

		if visited[hash] {
			continue
		}
		visited[hash] = true

		obj, err := repo.ReadObject(hash)
		if err != nil {
			t.Fatalf("failed to read object %s: %v", hash, err)
		}
		objectsRead++

		switch o := obj.(type) {
		case *StoredCommit:
			queue = append(queue, o.Tree)
			queue = append(queue, o.Parents...)

		case *StoredTree:
			for _, entry := range o.Entries {
				queue = append(queue, entry.ID)
			}

		case *StoredTag:
			queue = append(queue, o.Target)

		case *StoredBlob:

		default:
			t.Errorf("unexpected object type: %T", o)
		}
	}

	if objectsRead == 0 {
		t.Fatal("no objects were read from the repository")
	}
}
