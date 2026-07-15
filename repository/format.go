package repository

import (
	"errors"
	"fmt"
	"strings"

	"lindenii.org/go/furgit/config"
	"lindenii.org/go/furgit/object/id"
)

// maxFormatVersion is the highest repository format version furgit opens.
const maxFormatVersion = 1

// anyVersionExtensions are honored at every repository format version,
// for compatibility with repositories written before versioned extensions.
// Furgit acts on none of them.
//
//nolint:gochecknoglobals
var anyVersionExtensions = map[string]struct{}{
	"noop":            {},
	"preciousobjects": {},
	"partialclone":    {},
	"worktreeconfig":  {},
}

// version1Extensions require repository format version 1.
// Furgit acts on only some of them,
// and rejects those it cannot honor.
//
//nolint:gochecknoglobals
var version1Extensions = map[string]struct{}{
	"noop-v1":             {},
	"objectformat":        {},
	"compatobjectformat":  {},
	"refstorage":          {},
	"relativeworktrees":   {},
	"submodulepathconfig": {},
}

// ObjectFormat returns the object format of the repository.
func (repo *Repository) ObjectFormat() id.ObjectFormat {
	return repo.objectFormat
}

// detectObjectFormat validates the declared repository format
// and returns the object format that comes with it.
func detectObjectFormat(cfg *config.Config) (id.ObjectFormat, error) {
	version, err := detectFormatVersion(cfg)
	if err != nil {
		return id.ObjectFormatUnknown, err
	}

	objectFormat := id.ObjectFormatSHA1

	for _, entry := range cfg.Entries() {
		if entry.Section != "extensions" {
			continue
		}

		name := entry.Key
		if entry.Subsection != "" {
			name = entry.Subsection + "." + entry.Key
		}

		_, anyVersion := anyVersionExtensions[name]
		if anyVersion {
			continue
		}

		_, version1 := version1Extensions[name]
		if !version1 {
			if version == 0 {
				continue
			}

			return id.ObjectFormatUnknown, fmt.Errorf(
				"%w: unknown extension %q", ErrConfig, name,
			)
		}

		if version < 1 {
			return id.ObjectFormatUnknown, fmt.Errorf(
				"%w: extension %q requires repository format version 1",
				ErrConfig, name,
			)
		}

		switch name {
		case "objectformat":
			objectFormat, err = id.ParseObjectFormat(entry.Value)
			if err != nil {
				return id.ObjectFormatUnknown, fmt.Errorf(
					"%w: object format %q: %w", ErrConfig, entry.Value, err,
				)
			}
		case "refstorage":
			err = checkRefStorage(entry.Value)
			if err != nil {
				return id.ObjectFormatUnknown, err
			}
		case "compatobjectformat":
			return id.ObjectFormatUnknown, fmt.Errorf(
				"%w: unsupported extension %q", ErrConfig, name,
			)
		// TODO: properly handle the other extensions
		}
	}

	return objectFormat, nil
}

// detectFormatVersion returns the declared repository format version,
// which is zero when undeclared.
func detectFormatVersion(cfg *config.Config) (int, error) {
	version, err := cfg.Lookup("core", "", "repositoryformatversion").Int()
	if errors.Is(err, config.ErrMissing) {
		return 0, nil
	}

	if err != nil {
		return 0, fmt.Errorf("%w: repository format version: %w", ErrConfig, err)
	}

	if version < 0 || version > maxFormatVersion {
		return 0, fmt.Errorf(
			"%w: unsupported repository format version %d", ErrConfig, version,
		)
	}

	return version, nil
}

func checkRefStorage(value string) error {
	name, _, found := strings.Cut(value, "://")
	if !found {
		name = value
	}

	if name != "files" {
		return fmt.Errorf("%w: unsupported reference storage %q", ErrConfig, value)
	}

	// TODO: reftable

	return nil
}
