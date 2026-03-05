package repository

import (
	"fmt"
	"os"

	"codeberg.org/lindenii/furgit/config"
)

func parseRepositoryConfig(root *os.Root) (*config.Config, error) {
	configFile, err := root.Open("config")
	if err != nil {
		return nil, fmt.Errorf("repository: open config: %w", err)
	}

	defer func() { _ = configFile.Close() }()

	cfg, err := config.ParseConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("repository: parse config: %w", err)
	}

	return cfg, nil
}

// Config returns the parsed repository configuration snapshot.
//
// The returned pointer is owned by Repository. Callers should treat it as
// read-only.
func (repo *Repository) Config() *config.Config {
	return repo.config
}
