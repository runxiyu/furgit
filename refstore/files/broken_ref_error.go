package files

import "fmt"

type brokenRefError struct {
	name string
	err  error
}

func (err brokenRefError) Error() string {
	return fmt.Sprintf("refstore/files: broken reference %q: %v", err.name, err.err)
}

func (err brokenRefError) Unwrap() error {
	return err.err
}
