package testgit

import (
	"os/exec"
	"strings"
	"testing"
)

type RefFormatOptions struct {
	AllowOneLevel  bool
	RefspecPattern bool
}

func (repo *Repo) CheckRefFormat(tb testing.TB, name string, opts RefFormatOptions) error {
	tb.Helper()

	_, err := repo.run(tb, nil, "git", refFormatArgs("check-ref-format", name, opts)...)

	return err
}

func (repo *Repo) NormalizeRefFormat(tb testing.TB, name string, opts RefFormatOptions) (string, error) {
	tb.Helper()

	out, err := repo.run(tb, nil, "git", refFormatArgs("check-ref-format", name, opts, "--normalize")...)
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(string(out), "\n"), nil
}

func (repo *Repo) CheckBranchName(tb testing.TB, name string) (string, error) {
	tb.Helper()

	out, err := repo.run(tb, nil, "git", "check-ref-format", "--branch", name)
	if err != nil {
		return "", err
	}

	branchName := strings.TrimSuffix(string(out), "\n")
	if strings.HasPrefix(branchName, "refs/") {
		return branchName, nil
	}

	return "refs/heads/" + branchName, nil
}

func (repo *Repo) CheckTagName(tb testing.TB, name string) (string, error) {
	tb.Helper()

	if strings.HasPrefix(name, "-") || name == "HEAD" {
		return "", exec.ErrNotFound
	}

	_, err := repo.run(tb, nil, "git", "check-ref-format", "refs/tags/"+name)
	if err != nil {
		return "", err
	}

	return "refs/tags/" + name, nil
}

func refFormatArgs(command string, name string, opts RefFormatOptions, extra ...string) []string {
	args := []string{command}
	args = append(args, extra...)

	if opts.AllowOneLevel {
		args = append(args, "--allow-onelevel")
	}

	if opts.RefspecPattern {
		args = append(args, "--refspec-pattern")
	}

	return append(args, name)
}
