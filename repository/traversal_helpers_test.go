package repository_test

import (
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/repository"
)

func traverseTreeIter(repo *repository.Repository, root objectid.ObjectID) (int, error) {
	stack := []objectid.ObjectID{root}
	total := 0

	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		stored, err := repo.ReadStored(id)
		if err != nil {
			return 0, err
		}
		total++

		tree, ok := stored.Object().(*object.Tree)
		if !ok {
			continue
		}
		for i := len(tree.Entries) - 1; i >= 0; i-- {
			entry := tree.Entries[i]
			if entry.Mode == object.FileModeGitlink {
				continue
			}
			stack = append(stack, entry.ID)
		}
	}

	return total, nil
}

func traverseReachableIter(repo *repository.Repository, root objectid.ObjectID) (int, error) {
	stack := []objectid.ObjectID{root}
	visited := make(map[objectid.ObjectID]struct{})
	total := 0

	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if _, ok := visited[id]; ok {
			continue
		}
		visited[id] = struct{}{}

		stored, err := repo.ReadStored(id)
		if err != nil {
			return 0, err
		}
		total++

		switch obj := stored.Object().(type) {
		case *object.Commit:
			stack = append(stack, obj.Tree)
			stack = append(stack, obj.Parents...)
		case *object.Tree:
			for i := len(obj.Entries) - 1; i >= 0; i-- {
				entry := obj.Entries[i]
				if entry.Mode == object.FileModeGitlink {
					continue
				}
				stack = append(stack, entry.ID)
			}
		case *object.Tag:
			stack = append(stack, obj.Target)
		case *object.Blob:
		default:
			// Unknown parsed object variants are treated as leaves.
		}
	}

	return total, nil
}
