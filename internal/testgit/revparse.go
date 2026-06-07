package testgit

import (
	"fmt"
	"strings"
	"testing"

	"lindenii.org/go/furgit/object/id"
)

// RevParse resolves one revision spec to an object ID.
func (repo *Repo) RevParse(tb testing.TB, spec string) (id.ObjectID, error) {
	tb.Helper()

	stdout, err := repo.run(tb, nil, "git", "rev-parse", "--verify", "--end-of-options", spec)
	if err != nil {
		return id.ObjectID{}, fmt.Errorf("rev-parse %q: %w", spec, err)
	}

	objectID, err := repo.objectFormat.FromString(strings.TrimSuffix(string(stdout), "\n"))
	if err != nil {
		return id.ObjectID{}, fmt.Errorf("parse git rev-parse output %q: %w", string(stdout), err)
	}

	return objectID, nil
}
