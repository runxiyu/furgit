package refname

import "strings"

// Tag checks one tag shorthand and returns its fully-qualified
// refs/tags/... name.
func Tag(name string) (string, error) {
	if strings.HasPrefix(name, "-") || name == "HEAD" {
		return "", &NameError{Name: name, Reason: "invalid tag name"}
	}

	full := "refs/tags/" + name

	err := validate(full, 0)
	if err != nil {
		return "", err
	}

	return full, nil
}
