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
// for compatibility with repositories written before extensions were gated.
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

	err = checkExtensions(cfg, version)
	if err != nil {
		return id.ObjectFormatUnknown, err
	}

	err = checkRefStorage(cfg)
	if err != nil {
		return id.ObjectFormatUnknown, err
	}

	return lookupObjectFormat(cfg)
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

// checkExtensions rejects every declared extension
// that furgit cannot honor at version.
//
// It reads names rather than values,
// so a name declared more than once needs no resolving:
// every declaration of it reaches the same verdict.
func checkExtensions(cfg *config.Config, version int) error {
	for _, entry := range cfg.Entries() {
		if entry.Section != "extensions" {
			continue
		}

		name := extensionName(entry)

		_, anyVersion := anyVersionExtensions[name]
		if anyVersion {
			continue
		}

		_, version1 := version1Extensions[name]
		if !version1 {
			// Version 0 predates extensions,
			// so a name it does not know carries no meaning.
			if version == 0 {
				continue
			}

			return fmt.Errorf("%w: unknown extension %q", ErrConfig, name)
		}

		if version == 0 {
			return fmt.Errorf(
				"%w: extension %q requires repository format version 1",
				ErrConfig, name,
			)
		}

		if name == "compatobjectformat" {
			return fmt.Errorf("%w: unsupported extension %q", ErrConfig, name)
		}
	}

	return checkExtensionValues(cfg)
}

// checkExtensionValues validates the effective values of extensions
// whose names carry a value type.
func checkExtensionValues(cfg *config.Config) error {
	boolean := []string{
		"preciousobjects",
		"worktreeconfig",
		"relativeworktrees",
		"submodulepathconfig",
	}

	for _, name := range boolean {
		result := cfg.Lookup("extensions", "", name)
		if result.Kind == config.KindMissing {
			continue
		}

		_, err := result.Bool()
		if err != nil {
			return fmt.Errorf("%w: extension %q: %w", ErrConfig, name, err)
		}
	}

	result := cfg.Lookup("extensions", "", "partialclone")
	if result.Kind == config.KindMissing {
		return nil
	}

	_, err := result.String()
	if err != nil {
		return fmt.Errorf("%w: extension %q: %w", ErrConfig, "partialclone", err)
	}

	return nil
}

// extensionName returns the name one entry declares,
// without its section.
func extensionName(entry config.Entry) string {
	if entry.Subsection == "" {
		return entry.Key
	}

	return entry.Subsection + "." + entry.Key
}

// checkRefStorage rejects unimplemented reference storage formats.
func checkRefStorage(cfg *config.Config) error {
	result := cfg.Lookup("extensions", "", "refstorage")
	if result.Kind == config.KindMissing {
		return nil
	}

	value, err := result.String()
	if err != nil {
		return fmt.Errorf("%w: reference storage: %w", ErrConfig, err)
	}

	name, _, hasPayload := strings.Cut(value, "://")
	if hasPayload {
		return fmt.Errorf("%w: unsupported reference storage URI %q", ErrConfig, value)
	}

	if name != "files" {
		return fmt.Errorf("%w: unsupported reference storage %q", ErrConfig, value)
	}

	// TODO: reftable
	return nil
}

// lookupObjectFormat returns the declared object format,
// which is SHA-1 when undeclared.
func lookupObjectFormat(cfg *config.Config) (id.ObjectFormat, error) {
	result := cfg.Lookup("extensions", "", "objectformat")
	if result.Kind == config.KindMissing {
		return id.ObjectFormatSHA1, nil
	}

	value, err := result.String()
	if err != nil {
		return id.ObjectFormatUnknown, fmt.Errorf("%w: object format: %w", ErrConfig, err)
	}

	objectFormat, err := id.ParseObjectFormat(value)
	if err != nil {
		return id.ObjectFormatUnknown, fmt.Errorf(
			"%w: object format %q: %w", ErrConfig, value, err,
		)
	}

	return objectFormat, nil
}
