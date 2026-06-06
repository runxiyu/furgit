package name

import (
	"fmt"
	"strings"
)

// Branch checks a branch shorthand
// and returns its fully-qualified refs/heads/... name.
//
// Unlike Git in-repository branch parsing,
// this helper does not expand @{-n}.
func Branch(name string) (string, error) {
	full := "refs/heads/" + name
	if strings.HasPrefix(name, "-") || full == "refs/heads/HEAD" {
		return "", fmt.Errorf("%w: invalid branch name", ErrInvalidName)
	}

	err := validate(full, 0)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(name, "refs/") {
		return name, nil
	}

	return full, nil
}

// Tag checks a tag shorthand
// and returns its fully-qualified refs/tags/... name.
func Tag(name string) (string, error) {
	if strings.HasPrefix(name, "-") || name == "HEAD" {
		return "", fmt.Errorf("%w: invalid tag name", ErrInvalidName)
	}

	full := "refs/tags/" + name

	err := validate(full, 0)
	if err != nil {
		return "", err
	}

	return full, nil
}
