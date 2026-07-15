package repository

import (
	"errors"
	"fmt"
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

// Config returns the repository configuration as parsed by [Open].
//
// Later changes to the configuration file are not reflected.
//
// Labels: Life-Parent, Mut-No.
func (repo *Repository) Config() *config.Config {
	return repo.config
}
