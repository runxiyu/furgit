package main

import (
	"fmt"
	"os"
	"path/filepath"

	"codeberg.org/lindenii/furgit/format/packfile/ingest"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstore "codeberg.org/lindenii/furgit/object/store"
	"codeberg.org/lindenii/furgit/repository"
)

func run(repoPath, destinationPath, objectFormat string, fixThin, writeRev bool) error {
	var (
		algo objectid.Algorithm
		base objectstore.ReadingStore
		repo *repository.Repository
	)

	if repoPath != "" {
		repoRoot, err := os.OpenRoot(repoPath)
		if err != nil {
			return fmt.Errorf("open repo root: %w", err)
		}

		defer func() { _ = repoRoot.Close() }()

		repo, err = repository.Open(repoRoot)
		if err != nil {
			return fmt.Errorf("open repository: %w", err)
		}

		defer func() { _ = repo.Close() }()
	}

	algo, err := resolveAlgorithm(repo, objectFormat)
	if err != nil {
		return err
	}

	if fixThin {
		if repo == nil {
			return fmt.Errorf("fix-thin requires -r <repo>")
		}

		if repo.Algorithm() != algo {
			return fmt.Errorf("algorithm mismatch: repo=%s flag=%s", repo.Algorithm(), algo)
		}

		base = repo.Objects()
	}

	absDestination, err := filepath.Abs(destinationPath)
	if err != nil {
		return fmt.Errorf("absolute destination path: %w", err)
	}

	destinationRoot, err := os.OpenRoot(absDestination)
	if err != nil {
		return fmt.Errorf("open destination root: %w", err)
	}

	defer func() { _ = destinationRoot.Close() }()

	pending, err := ingest.Ingest(os.Stdin, algo, ingest.Options{
		FixThin:            fixThin,
		WriteRev:           writeRev,
		Base:               base,
		RequireTrailingEOF: true,
	})
	if err != nil {
		return err
	}

	if pending.Header().ObjectCount == 0 {
		discarded, err := pending.Discard()
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(os.Stdout, "pack\t%s\n", discarded.PackHash.String())

		return nil
	}

	result, err := pending.Continue(destinationRoot)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "pack\t%s\n", result.PackHash.String())

	return nil
}
