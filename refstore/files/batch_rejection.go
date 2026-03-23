package files

import (
	"errors"

	"codeberg.org/lindenii/furgit/refstore"
)

func isBatchRejected(err error) bool {
	return errors.Is(err, refstore.ErrReferenceNotFound) ||
		errors.As(err, new(*refstore.InvalidNameError)) ||
		errors.As(err, new(*refstore.InvalidValueError)) ||
		errors.As(err, new(*refstore.DuplicateUpdateError)) ||
		errors.As(err, new(*refstore.CreateExistsError)) ||
		errors.As(err, new(*refstore.IncorrectOldValueError)) ||
		errors.As(err, new(*refstore.ExpectedDetachedError)) ||
		errors.As(err, new(*refstore.ExpectedSymbolicError)) ||
		errors.As(err, new(*refstore.NameConflictError))
}
