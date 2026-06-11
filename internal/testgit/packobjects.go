package testgit

import (
	"bytes"
	"iter"
	"path/filepath"
	"strings"
	"testing"

	"lindenii.org/go/furgit/object/id"
)

// PackObjectsOptions controls one pack-objects invocation.
type PackObjectsOptions struct {
	// RevIndex requests writing a .rev reverse index alongside the pack.
	RevIndex bool
}

// PackObjects packs the supplied objects with git pack-objects
// into a temporary directory,
// and returns the artifact path prefix "<dir>/pack-<hash>",
// to which ".pack", ".idx", and ".rev" suffixes apply.
func (repo *Repo) PackObjects(tb testing.TB, oids iter.Seq[id.ObjectID], opts PackObjectsOptions) (string, error) {
	tb.Helper()

	dir := tb.TempDir()

	var stdin bytes.Buffer
	for oid := range oids {
		stdin.WriteString(oid.String())
		stdin.WriteByte('\n')
	}

	revIndex := "false"
	if opts.RevIndex {
		revIndex = "true"
	}

	out, err := repo.run(tb, &stdin,
		"git", "-c", "pack.writeReverseIndex="+revIndex,
		"pack-objects", "--end-of-options", filepath.Join(dir, "pack"))
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "pack-"+strings.TrimSpace(string(out))), nil
}
