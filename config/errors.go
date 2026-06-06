package config

import "fmt"

// ParseError describes a syntactic error in Git config input.
type ParseError struct {
	Line   int
	Reason string
}

func (err *ParseError) Error() string {
	if err.Line > 0 {
		return fmt.Sprintf("config: parse line %d: %s", err.Line, err.Reason)
	}

	return "config: parse: " + err.Reason
}

// LookupError describes an invalid lookup result conversion.
type LookupError struct {
	Kind      Kind
	Operation string
}

func (err *LookupError) Error() string {
	switch err.Kind {
	case ValueMissing:
		return fmt.Sprintf("config: %s: missing config value", err.Operation)
	case ValueValueless:
		return fmt.Sprintf("config: %s: valueless config key", err.Operation)
	case ValueString:
		return fmt.Sprintf("config: %s: invalid string config value", err.Operation)
	default:
		return fmt.Sprintf("config: %s: unknown value kind %d", err.Operation, err.Kind)
	}
}

// ValueError describes a typed value conversion failure.
type ValueError struct {
	Operation string
	Value     string
	Reason    string
	Err       error
}

func (err *ValueError) Error() string {
	if err.Err != nil {
		return fmt.Sprintf("config: %s %q: %s: %v", err.Operation, err.Value, err.Reason, err.Err)
	}

	return fmt.Sprintf("config: %s %q: %s", err.Operation, err.Value, err.Reason)
}

func (err *ValueError) Unwrap() error {
	return err.Err
}
