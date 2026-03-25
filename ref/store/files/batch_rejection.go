package files

import (
	"errors"

	refstore "codeberg.org/lindenii/furgit/ref/store"
)

func isBatchRejected(err error) bool {
	_, invalidName := errors.AsType[*refstore.InvalidNameError](err)
	_, invalidValue := errors.AsType[*refstore.InvalidValueError](err)
	_, duplicateUpdate := errors.AsType[*refstore.DuplicateUpdateError](err)
	_, createExists := errors.AsType[*refstore.CreateExistsError](err)
	_, incorrectOldValue := errors.AsType[*refstore.IncorrectOldValueError](err)
	_, expectedDetached := errors.AsType[*refstore.ExpectedDetachedError](err)
	_, expectedSymbolic := errors.AsType[*refstore.ExpectedSymbolicError](err)
	_, nameConflict := errors.AsType[*refstore.NameConflictError](err)

	return errors.Is(err, refstore.ErrReferenceNotFound) ||
		invalidName ||
		invalidValue ||
		duplicateUpdate ||
		createExists ||
		incorrectOldValue ||
		expectedDetached ||
		expectedSymbolic ||
		nameConflict
}
