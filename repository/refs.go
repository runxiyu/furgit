package repository

import (
	"errors"
	"fmt"
	"time"

	"lindenii.org/go/furgit/config"
	"lindenii.org/go/furgit/ref/store"
	"lindenii.org/go/furgit/ref/store/files"
)

// Refs returns the reference store of the repository.
//
// Labels: Life-Parent.
func (repo *Repository) Refs() interface {
	store.Reader
	store.Transactioner
	store.Batcher
} {
	return repo.refs
}

func detectRefOptions(cfg *config.Config) (files.Options, error) {
	looseLockTimeout, err := detectLockTimeout(
		cfg, "filesreflocktimeout", files.DefaultLooseLockTimeout,
	)
	if err != nil {
		return files.Options{}, err
	}

	packedLockTimeout, err := detectLockTimeout(
		cfg, "packedrefstimeout", files.DefaultPackedLockTimeout,
	)
	if err != nil {
		return files.Options{}, err
	}

	return files.Options{
		LooseLockTimeout:  looseLockTimeout,
		PackedLockTimeout: packedLockTimeout,
		Fsync:             files.FsyncNone,
		// TODO: respect fsync options
	}, nil
}

func detectLockTimeout(cfg *config.Config, key string, fallback time.Duration) (time.Duration, error) {
	milliseconds, err := cfg.Lookup("core", "", key).Int()
	if errors.Is(err, config.ErrMissing) {
		return fallback, nil
	}

	if err != nil {
		return 0, fmt.Errorf("%w: core.%s: %w", ErrConfig, key, err)
	}

	return time.Duration(milliseconds) * time.Millisecond, nil
}
