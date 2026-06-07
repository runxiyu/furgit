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

	// Sign requests a signed tag via git tag -s,
	// using the gpg.format and user.signingkey configured on the repo.
	Sign bool
}

// TagAnnotated creates an annotated tag object and returns its object ID.
func (repo *Repo) TagAnnotated(
	tb testing.TB,
	name string,
	target id.ObjectID,
	opts TagAnnotatedOptions,
) (id.ObjectID, error) {
	tb.Helper()

	args := []string{"tag", "-a"}
	if opts.Sign {
		args = append(args, "-s")
	}

	args = append(args, "-m", opts.Message, "--end-of-options", name, target.String())

	cmd := repo.command(tb, "git", args...)

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
