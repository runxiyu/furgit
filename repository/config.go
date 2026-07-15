package repository

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"lindenii.org/go/furgit/config"
)

// ErrConfig indicates that a repository's configuration
// is erroneous or unsupported.
var ErrConfig = errors.New("repository: erroneous or unsupported repository configuration")

// parseConfig parses the configuration file of the repository.
func parseConfig(commonRoot *os.Root) (*config.Config, error) {
	file, err := commonRoot.Open("config")
	if err != nil {
		return nil, fmt.Errorf("repository: open config: %w", err)
	}

	defer func() { _ = file.Close() }()

	cfg, err := config.Parse(file)
	if err != nil {
		return nil, fmt.Errorf("repository: parse config: %w", err)
	}

	return cfg, nil
}

// effectiveConfig returns the configuration in force for one worktree.
//
// It extends the common configuration with the per-worktree file
// when the repository declares that it keeps one,
// which overrides the common configuration.
//
// The repository format is not read from the per-worktree file,
// and so is not resolved from the result.
func effectiveConfig(gitRoot *os.Root, common *config.Config) (*config.Config, error) {
	result := common.Lookup("extensions", "", "worktreeconfig")
	if result.Kind == config.KindMissing {
		return common, nil
	}

	declared, err := result.Bool()
	if err != nil {
		return nil, fmt.Errorf("%w: worktree configuration: %w", ErrConfig, err)
	}

	if !declared {
		return common, nil
	}

	file, err := gitRoot.Open("config.worktree")
	if errors.Is(err, fs.ErrNotExist) {
		return common, nil
	}

	if err != nil {
		return nil, fmt.Errorf("repository: open config.worktree: %w", err)
	}

	defer func() { _ = file.Close() }()

	worktree, err := config.Parse(file)
	if err != nil {
		return nil, fmt.Errorf("repository: parse config.worktree: %w", err)
	}

	return config.Merge(common, worktree), nil
}

// Config returns the repository configuration as parsed by [Open].
//
// Later changes to the configuration file are not reflected.
//
// Labels: Life-Parent, Mut-No.
func (repo *Repository) Config() *config.Config {
	return repo.config
}
