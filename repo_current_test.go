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

	headRef, err := repo.ResolveHead()
	if err != nil {
		t.Fatalf("failed to resolve HEAD: %v", err)
	}

	var headHash Hash

	switch headRef.Kind {
	case HeadKindDetached:
		headHash = headRef.Hash
	case HeadKindSymbolic:
		headHash, err = repo.ResolveRef(headRef.Ref)
		if err != nil {
			t.Fatalf("failed to resolve symbolic HEAD ref %v: %v", headRef, err)
		}
	default:
		t.Fatalf("unexpected HEAD ref kind: %v", headRef.Kind)
	}

	visited := make(map[Hash]bool)
	var visitQueue []Hash
	visitQueue = append(visitQueue, headHash)

	objectsRead := 0
	errors := 0
	for len(visitQueue) > 0 {
		hash := visitQueue[0]
		visitQueue = visitQueue[1:]

		if visited[hash] {
			continue
		}
		visited[hash] = true

		obj, err := repo.ReadObject(hash)
		if err != nil {
			t.Logf("failed to read object %s: %v", hash, err)
			errors++
			if errors > 10 {
				t.Fatalf("too many errors (%d) reading objects", errors)
			}
			continue
		}
		objectsRead++

		switch o := obj.(type) {
		case *StoredCommit:
			visitQueue = append(visitQueue, o.Tree)
			visitQueue = append(visitQueue, o.Parents...)
		case *StoredTree:
			for _, entry := range o.Entries {
				visitQueue = append(visitQueue, entry.ID)
			}
		case *StoredTag:
			visitQueue = append(visitQueue, o.Target)
		case *StoredBlob:
		default:
			t.Errorf("unexpected object type: %T", o)
		}
	}

	if objectsRead == 0 {
		t.Fatal("no objects were read from the repository")
	}

	t.Logf("Read %d objects from current repository HEAD (%d errors)", objectsRead, errors)
	if errors > 0 {
		t.Fatalf("encountered %d errors during enumeration", errors)
	}
}
