package testgit

import (
	"fmt"
	"strings"
	"testing"

	"lindenii.org/go/furgit/object/id"
)

// CommitTreeOptions configures [Repo.CommitTree].
//
//exhaustruct:optional
type CommitTreeOptions struct {
	Message       string
	Author        Identity
	Committer     Identity
	AuthorDate    string
	CommitterDate string

	// Sign requests a signed commit via git commit-tree -S,
	// using the gpg.format and user.signingkey configured on the repo.
	Sign bool
}

// CommitTree creates a commit object from a tree and optional parents,
// and returns its object ID.
func (repo *Repo) CommitTree(
	tb testing.TB,
	tree id.ObjectID,
	opts CommitTreeOptions,
	parents ...id.ObjectID,
) (id.ObjectID, error) {
	tb.Helper()

	args := make([]string, 0, 1+2*len(parents)+4)
	args = append(args, "commit-tree")

	for _, parent := range parents {
		args = append(args, "-p", parent.String())
	}

	if opts.Sign {
		args = append(args, "-S")
	}

	args = append(args, "-m", opts.Message, "--end-of-options", tree.String())

	cmd := repo.command(tb, "git", args...)
	if opts.Author.Name != "" {
		cmd.Env = setEnv(cmd.Env, "GIT_AUTHOR_NAME", opts.Author.Name)
	}

	if opts.Author.Email != "" {
		cmd.Env = setEnv(cmd.Env, "GIT_AUTHOR_EMAIL", opts.Author.Email)
	}

	if opts.AuthorDate != "" {
		cmd.Env = setEnv(cmd.Env, "GIT_AUTHOR_DATE", opts.AuthorDate)
	}

	if opts.Committer.Name != "" {
		cmd.Env = setEnv(cmd.Env, "GIT_COMMITTER_NAME", opts.Committer.Name)
	}

	if opts.Committer.Email != "" {
		cmd.Env = setEnv(cmd.Env, "GIT_COMMITTER_EMAIL", opts.Committer.Email)
	}

	if opts.CommitterDate != "" {
		cmd.Env = setEnv(cmd.Env, "GIT_COMMITTER_DATE", opts.CommitterDate)
	}

	stdout, err := cmd.Output()
	if err != nil {
		return id.ObjectID{}, fmt.Errorf("commit-tree: %w", err)
	}

	commitID, err := repo.objectFormat.FromString(strings.TrimSuffix(string(stdout), "\n"))
	if err != nil {
		return id.ObjectID{}, fmt.Errorf("parse git commit-tree output %q: %w", string(stdout), err)
	}

	return commitID, nil
}
