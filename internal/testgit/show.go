package testgit

import (
	"fmt"
	"testing"

	"lindenii.org/go/furgit/object/id"
)

// ShowFormat returns git-show output for one pretty format.
func (repo *Repo) ShowFormat(tb testing.TB, oid id.ObjectID, format string) ([]byte, error) {
	tb.Helper()

	stdout, err := repo.run(tb, nil, "git", "show", "--no-patch", "--no-color", "--format="+format, "--end-of-options", oid.String())
	if err != nil {
		return nil, fmt.Errorf("show --format=%q %s: %w", format, oid, err)
	}

	return stdout, nil
}
