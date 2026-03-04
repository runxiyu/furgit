package repository

import (
	"fmt"
	"os"

	"codeberg.org/lindenii/furgit/config"
	"codeberg.org/lindenii/furgit/objectid"
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

func detectObjectAlgorithm(cfg *config.Config) (objectid.Algorithm, error) {
	algoName := cfg.Lookup("extensions", "", "objectformat").Value
	if algoName == "" {
		algoName = objectid.AlgorithmSHA1.String()
	}

	algo, ok := objectid.ParseAlgorithm(algoName)
	if !ok {
		return objectid.AlgorithmUnknown, fmt.Errorf("repository: unsupported object format %q", algoName)
	}

	return algo, nil
}
