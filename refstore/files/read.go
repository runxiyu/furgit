package files

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/refstore"
)

type brokenRefError struct {
	name string
	err  error
}

func (err brokenRefError) Error() string {
	return fmt.Sprintf("refstore/files: broken reference %q: %v", err.name, err.err)
}

func (err brokenRefError) Unwrap() error {
	return err.err
}

func (store *Store) readLooseRef(name string) (ref.Ref, error) { //nolint:ireturn
	refPath := store.loosePath(name)

	data, err := store.rootFor(refPath.root).ReadFile(refPath.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, refstore.ErrReferenceNotFound
		}

		return nil, err
	}

	line := trimTrailingRefWhitespace(string(data))
	if strings.HasPrefix(line, "ref:") {
		target := strings.TrimLeftFunc(line[len("ref:"):], isRefWhitespace)
		if target == "" {
			return nil, brokenRefError{name: name, err: fmt.Errorf("empty symbolic target")}
		}

		return ref.Symbolic{
			RefName: name,
			Target:  target,
		}, nil
	}

	id, err := objectid.ParseHex(store.algo, line)
	if err != nil {
		return nil, brokenRefError{name: name, err: err}
	}

	return ref.Detached{
		RefName: name,
		ID:      id,
	}, nil
}

func trimTrailingRefWhitespace(text string) string {
	return strings.TrimRightFunc(text, isRefWhitespace)
}

func isRefWhitespace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	default:
		return false
	}
}
