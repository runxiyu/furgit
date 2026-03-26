package refname

import "strings"

// Validate checks whether name is one valid Git refname.
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
			return &NameError{Name: name, Reason: "single @ is not allowed"}
		}

		sanitized.WriteByte('-')
	}

	for {
		if sanitized != nil && sanitized.Len() > 0 {
			sanitized.WriteByte('/')
		}

		componentLen, err := checkRefnameComponent(remaining, &flags, sanitized, name)
		switch {
		case sanitized != nil && componentLen == 0:
		case componentLen <= 0:
			if err != nil {
				return err
			}

			return &NameError{Name: name, Reason: "component has zero length"}
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
			return &NameError{Name: name, Reason: "name ends with '.'"}
		}
	}

	if flags&refnameAllowOneLevel == 0 && componentCount < 2 {
		return &NameError{Name: name, Reason: "one-level refname is not allowed"}
	}

	return nil
}
