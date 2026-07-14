package testgit

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"lindenii.org/go/furgit/object/id"
)

// ErrInvalidPackObjectsOptions reports an inconsistent
// combination of pack-objects options.
var ErrInvalidPackObjectsOptions = errors.New("internal/testgit: invalid pack-objects options")

// PackObjectsOptions controls one on-disk pack-objects invocation.
//
//exhaustruct:optional
type PackObjectsOptions struct {
	// RevIndex requests writing a .rev reverse index alongside the pack.
	RevIndex bool

	// Revs changes how the input is interpreted.
	//
	// When false, each include is one object name,
	// and exactly those objects are packed.
	//
	// When true, each include is a revision argument,
	// and pack-objects packs the closure of objects
	// reachable from the includes but not from the excludes,
	// the same walk git rev-list --objects performs.
	Revs bool

	// Exclude lists objects to omit by reachability,
	// fed as "^<oid>" revision arguments.
	// A non-nil Exclude requires Revs.
	Exclude []id.ObjectID
}

// PackObjects packs the include objects with git pack-objects
// into a temporary directory,
// and returns the artifact path prefix "<dir>/pack-<hash>",
// to which ".pack", ".idx", and ".rev" suffixes apply.
func (repo *Repo) PackObjects(tb testing.TB, include []id.ObjectID, opts PackObjectsOptions) (string, error) {
	tb.Helper()

	if opts.Exclude != nil && !opts.Revs {
		return "", fmt.Errorf("%w: Exclude requires Revs", ErrInvalidPackObjectsOptions)
	}

	dir := tb.TempDir()

	revIndex := "false"
	if opts.RevIndex {
		revIndex = "true"
	}

	args := []string{"-c", "pack.writeReverseIndex=" + revIndex, "pack-objects"}
	if opts.Revs {
		args = append(args, "--revs")
	}

	args = append(args, "--end-of-options", filepath.Join(dir, "pack"))

	out, err := repo.run(tb, packObjectsStdin(include, opts.Exclude), "git", args...)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "pack-"+strings.TrimSpace(string(out))), nil
}

// PackObjectsStdoutOptions controls one streamed pack-objects invocation.
type PackObjectsStdoutOptions struct {
	// Revs changes how the input is interpreted.
	//
	// When false, each include is one object name,
	// and exactly those objects are packed.
	//
	// When true, each include is a revision argument,
	// and pack-objects packs the closure of objects
	// reachable from the includes but not from the excludes,
	// the same walk git rev-list --objects performs.
	Revs bool

	// Thin omits excluded base objects from the pack,
	// producing a thin pack that must be completed before use.
	// Thin requires Revs.
	Thin bool

	// Exclude lists objects to omit by reachability,
	// fed as "^<oid>" revision arguments.
	// A non-nil Exclude requires Revs.
	Exclude []id.ObjectID
}

// PackObjectsStdout packs the include objects with git pack-objects
// and returns the pack stream written to standard output.
func (repo *Repo) PackObjectsStdout(tb testing.TB, include []id.ObjectID, opts PackObjectsStdoutOptions) ([]byte, error) {
	tb.Helper()

	if opts.Exclude != nil && !opts.Revs {
		return nil, fmt.Errorf("%w: Exclude requires Revs", ErrInvalidPackObjectsOptions)
	}

	if opts.Thin && !opts.Revs {
		return nil, fmt.Errorf("%w: Thin requires Revs", ErrInvalidPackObjectsOptions)
	}

	args := []string{"pack-objects", "--stdout"}
	if opts.Revs {
		args = append(args, "--revs")
	}

	if opts.Thin {
		args = append(args, "--thin")
	}

	out, err := repo.run(tb, packObjectsStdin(include, opts.Exclude), "git", args...)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// packObjectsStdin builds the pack-objects standard input:
// the include object IDs,
// followed by the exclude object IDs as "^<oid>" lines.
func packObjectsStdin(include, exclude []id.ObjectID) *bytes.Buffer {
	var stdin bytes.Buffer

	for _, oid := range include {
		stdin.WriteString(oid.String())
		stdin.WriteByte('\n')
	}

	for _, oid := range exclude {
		stdin.WriteByte('^')
		stdin.WriteString(oid.String())
		stdin.WriteByte('\n')
	}

	return &stdin
}
