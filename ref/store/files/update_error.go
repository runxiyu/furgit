package files

import "fmt"

type updateContextError struct {
	name string
	err  error
}

func (err *updateContextError) Error() string {
	return fmt.Sprintf("refstore/files: update %q: %v", err.name, err.err)
}

func (err *updateContextError) Unwrap() error {
	if err == nil {
		return nil
	}

	return err.err
}

func wrapUpdateError(name string, err error) error {
	if err == nil || name == "" {
		return err
	}

	return &updateContextError{name: name, err: err}
}
