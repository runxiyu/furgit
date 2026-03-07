package refname

import "strings"

// Branch checks one branch shorthand and returns its fully-qualified
// refs/heads/... name.
//
// Unlike Git in-repository branch parsing, this helper does not expand @{-n}.
func Branch(name string) (string, error) {
	full := "refs/heads/" + name
	if strings.HasPrefix(name, "-") || full == "refs/heads/HEAD" {
		return "", &NameError{Name: name, Reason: "invalid branch name"}
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
