package repository_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/repository"
)

func TestRepositoryDepthFirstEnumerationFromHEAD(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repoHarness := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		_, _, commit1 := repoHarness.MakeCommit(t, "walk-one")
		blob2, tree2 := repoHarness.MakeSingleFileTree(t, "second.txt", []byte("second\n"))
		commit2 := repoHarness.CommitTree(t, tree2, "walk-two", commit1)
		_ = blob2
		repoHarness.UpdateRef(t, "refs/heads/main", commit2)
		repoHarness.SymbolicRef(t, "HEAD", "refs/heads/main")

		repo, err := repository.Open(repoHarness.Dir())
		if err != nil {
			t.Fatalf("repository.Open: %v", err)
		}
		defer func() { _ = repo.Close() }()

		head, err := repo.ResolveRefFully("HEAD")
		if err != nil {
			t.Fatalf("ResolveRefFully(HEAD): %v", err)
		}

		visited := make(map[objectid.ObjectID]bool)
		queue := []objectid.ObjectID{head.ID}
		objectsRead := 0

		for len(queue) > 0 {
			id := queue[0]
			queue = queue[1:]

			if visited[id] {
				continue
			}
			visited[id] = true

			stored, err := repo.ReadStored(id)
			if err != nil {
				t.Fatalf("ReadStored(%s): %v", id, err)
			}
			objectsRead++

			switch obj := stored.Object().(type) {
			case *object.Commit:
				queue = append(queue, obj.Tree)
				queue = append(queue, obj.Parents...)
			case *object.Tree:
				for _, entry := range obj.Entries {
					queue = append(queue, entry.ID)
				}
			case *object.Tag:
				queue = append(queue, obj.Target)
			case *object.Blob:
			default:
				t.Fatalf("unexpected object type: %T", obj)
			}
		}

		if objectsRead == 0 {
			t.Fatalf("no objects were enumerated from HEAD")
		}
	})
}
