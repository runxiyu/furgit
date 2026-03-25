package files

import "errors"

func batchResultError(err error) error {
	updateErr, ok := errors.AsType[*updateContextError](err)
	if ok {
		return updateErr.err
	}

	return err
}

func batchResultName(err error) string {
	updateErr, ok := errors.AsType[*updateContextError](err)
	if !ok {
		return ""
	}

	return updateErr.name
}
