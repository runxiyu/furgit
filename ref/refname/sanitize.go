package refname

import (
	"fmt"
	"strings"
)

// SanitizeComponent mutates component until it satisfies
// sanitize_refname_component.
func SanitizeComponent(component string) string {
	var builder strings.Builder

	err := checkOrSanitizeRefname(component, refnameAllowOneLevel, &builder)
	if err != nil {
		panic(fmt.Sprintf("ref: sanitize component %q: %v", component, err))
	}

	return builder.String()
}
