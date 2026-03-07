package refname

import "fmt"

// NameError reports one invalid reference name.
type NameError struct {
	Name   string
	Reason string
}

// Error implements error.
func (err *NameError) Error() string {
	return fmt.Sprintf("ref: invalid name %q: %s", err.Name, err.Reason)
}
