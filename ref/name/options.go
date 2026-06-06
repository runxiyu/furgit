package name

import "fmt"

const (
	nameAllowOneLevel = 1 << iota
	nameRefspecPattern
)

// Options controls Git name validation.
type Options struct {
	// AllowOneLevel permits one-component names like HEAD.
	AllowOneLevel bool

	// RefspecPattern permits a '*' anywhere in the name.
	RefspecPattern bool
}

// String returns a text form of the options.
//
// This is mainly just for tests.
func (options Options) String() string {
	return fmt.Sprintf("allow_onelevel=%t,refspec_pattern=%t", options.AllowOneLevel, options.RefspecPattern)
}

func (options Options) flags() int {
	var flags int
	if options.AllowOneLevel {
		flags |= nameAllowOneLevel
	}

	if options.RefspecPattern {
		flags |= nameRefspecPattern
	}

	return flags
}
