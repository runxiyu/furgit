package files

import (
	"errors"
	"strings"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/ref/refname"
	"codeberg.org/lindenii/furgit/refstore"
)

func isBatchRejected(err error) bool {
	var nameErr *refname.NameError

	if errors.As(err, &nameErr) {
		return true
	}

	if errors.Is(err, objectid.ErrInvalidAlgorithm) || errors.Is(err, refstore.ErrReferenceNotFound) {
		return true
	}

	msg := err.Error()

	return strings.Contains(msg, "empty reference name") ||
		strings.Contains(msg, "empty symbolic target") ||
		strings.Contains(msg, "empty symbolic old target") ||
		strings.Contains(msg, "already exists") ||
		strings.Contains(msg, "is missing") ||
		strings.Contains(msg, "is not detached") ||
		strings.Contains(msg, "is not symbolic") ||
		strings.Contains(msg, "expected") ||
		strings.Contains(msg, "reference name conflict") ||
		strings.Contains(msg, "non-empty directory blocks reference")
}
