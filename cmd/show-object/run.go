package main

import (
	"fmt"
	"os"

	"codeberg.org/lindenii/furgit/repository"
)

func run(repoPath, name *string) error {
	root, err := os.OpenRoot(*repoPath)
	if err != nil {
		return fmt.Errorf("open repo root: %w", err)
	}

	defer func() { _ = root.Close() }()

	repo, err := repository.Open(root)
	if err != nil {
		return fmt.Errorf("open repository: %w", err)
	}

	id, err := resolveInput(repo, *name)
	if err != nil {
		_ = repo.Close()

		return fmt.Errorf("resolve %q: %w", *name, err)
	}

	s, err := repo.Fetcher().ExactObject(id)
	if err != nil {
		_ = repo.Close()

		return fmt.Errorf("read object %s: %w", id, err)
	}

	printStored(s)

	err = repo.Close()
	if err != nil {
		return fmt.Errorf("close repository: %w", err)
	}

	return nil
}
