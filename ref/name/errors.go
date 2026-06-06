package name

import "fmt"

// NameError reports an invalid reference name.
type NameError struct {
	Name   string
	Reason string
}

// Error implements error.
func (err *NameError) Error() string {
	return fmt.Sprintf("ref/name: invalid name %q: %s", err.Name, err.Reason)
}
