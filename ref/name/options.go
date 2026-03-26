package refname

import "fmt"

// Options controls Git refname validation.
type Options struct {
	// AllowOneLevel permits one-component refnames like HEAD.
	AllowOneLevel bool

	// RefspecPattern permits one '*' anywhere in the refname.
	RefspecPattern bool
}

// String returns one stable text form of the options.
func (options Options) String() string {
	return fmt.Sprintf("allow_onelevel=%t,refspec_pattern=%t", options.AllowOneLevel, options.RefspecPattern)
}

func (options Options) flags() int {
	var flags int
	if options.AllowOneLevel {
		flags |= refnameAllowOneLevel
	}

	if options.RefspecPattern {
		flags |= refnameRefspecPattern
	}

	return flags
}
