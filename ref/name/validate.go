package name

import (
	"fmt"
	"strings"
)

// Validate checks whether name is a valid Git name.
func Validate(name string, options Options) error {
	return validate(name, options.flags())
}

func validate(name string, flags int) error {
	return checkOrSanitizeRefname(name, flags, nil)
}

func checkOrSanitizeRefname(name string, flags int, sanitized *strings.Builder) error {
	componentCount := 0
	remaining := name

	if name == "@" {
		if sanitized == nil {
			return fmt.Errorf("%w: single @ is not allowed", ErrInvalidName)
		}

		sanitized.WriteByte('-')
	}

	for {
		if sanitized != nil && sanitized.Len() > 0 {
			sanitized.WriteByte('/')
		}

		componentLen, err := checkRefnameComponent(remaining, &flags, sanitized)
		switch {
		case sanitized != nil && componentLen == 0:
		case componentLen <= 0:
			if err != nil {
				return err
			}

			return fmt.Errorf("%w: component has zero length", ErrInvalidName)
		case err != nil:
			return err
		}

		componentCount++

		if componentLen == len(remaining) {
			break
		}

		remaining = remaining[componentLen+1:]
	}

	componentLen := len(remaining)
	if componentLen > 0 && remaining[componentLen-1] == '.' {
		if sanitized == nil {
			return fmt.Errorf("%w: name ends with '.'", ErrInvalidName)
		}
	}

	if flags&nameAllowOneLevel == 0 && componentCount < 2 {
		return fmt.Errorf("%w: one-level name is not allowed", ErrInvalidName)
	}

	return nil
}

// SanitizeComponent sanitizes components.
func SanitizeComponent(component string) string {
	var builder strings.Builder

	err := checkOrSanitizeRefname(component, nameAllowOneLevel, &builder)
	if err != nil {
		panic(fmt.Sprintf("ref/name: sanitize component %q: %v", component, err))
	}

	return builder.String()
}
