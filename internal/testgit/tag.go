package testgit

import (
	"fmt"
	"testing"

	"lindenii.org/go/furgit/object/id"
)

// TagAnnotatedOptions configures [Repo.TagAnnotated].
type TagAnnotatedOptions struct {
	Message    string
	Tagger     Identity
	TaggerDate string
}

// TagAnnotated creates an annotated tag object and returns its object ID.
func (repo *Repo) TagAnnotated(
	tb testing.TB,
	name string,
	target id.ObjectID,
	opts TagAnnotatedOptions,
) (id.ObjectID, error) {
	tb.Helper()

	cmd := repo.command(tb, "git", "tag", "-a", "-m", opts.Message, "--end-of-options", name, target.String())

	if opts.Tagger.Name != "" {
		cmd.Env = setEnv(cmd.Env, "GIT_COMMITTER_NAME", opts.Tagger.Name)
	}

	if opts.Tagger.Email != "" {
		cmd.Env = setEnv(cmd.Env, "GIT_COMMITTER_EMAIL", opts.Tagger.Email)
	}

	if opts.TaggerDate != "" {
		cmd.Env = setEnv(cmd.Env, "GIT_COMMITTER_DATE", opts.TaggerDate)
	}

	err := cmd.Run()
	if err != nil {
		return id.ObjectID{}, fmt.Errorf("tag -a %q %s: %w", name, target, err)
	}

	return repo.RevParse(tb, "refs/tags/"+name)
}
