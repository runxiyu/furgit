package testgit

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
)

// HashObject hashes and writes an object,
// and returns its object ID.
func (repo *Repo) HashObject(tb testing.TB, ty typ.Type, body io.Reader) (id.ObjectID, error) {
	tb.Helper()

	stdout, err := repo.Run(tb, body, "git", "hash-object", "-t", ty.Name(), "-w", "--stdin", "--literally")
	if err != nil {
		return id.ObjectID{}, fmt.Errorf("hash-object: %w", err)
	}

	objectID, err := repo.objectFormat.FromString(strings.TrimSuffix(string(stdout), "\n"))
	if err != nil {
		return id.ObjectID{}, fmt.Errorf("parse git hash-object output %q: %w", string(stdout), err)
	}

	return objectID, nil
}
