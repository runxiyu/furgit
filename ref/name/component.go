package name

import "strings"

const lockSuffix = ".lock"

func nameDisposition(ch byte) byte {
	switch {
	case ch == '/':
		return 1
	case ch == '.':
		return 2
	case ch == '{':
		return 3
	case ch == '*':
		return 5
	case ch < 0x20 || ch == 0x7f:
		return 4
	case ch == ':' || ch == '?' || ch == '[' || ch == '\\' || ch == '^' || ch == '~' || ch == ' ' || ch == '\t':
		return 4
	default:
		return 0
	}
}

func checkRefnameComponent(name string, flags *int, sanitized *strings.Builder, fullName string) (int, error) {
	var last byte

	var componentStart int
	if sanitized != nil {
		componentStart = sanitized.Len()
	}

	for i := range len(name) {
		ch := name[i]
		disp := nameDisposition(ch)

		if sanitized != nil && disp != 1 {
			sanitized.WriteByte(ch)
		}

		switch disp {
		case 1:
			goto out
		case 2:
			if last == '.' {
				if sanitized != nil {
					truncateBuilder(sanitized, sanitized.Len()-1)
				} else {
					return 0, &NameError{Name: fullName, Reason: "name contains '..'"}
				}
			}
		case 3:
			if last == '@' {
				if sanitized != nil {
					overwriteBuilderAt(sanitized, sanitized.Len()-1)
				} else {
					return 0, &NameError{Name: fullName, Reason: "name contains '@{'"}
				}
			}
		case 4:
			if sanitized != nil {
				overwriteBuilderAt(sanitized, sanitized.Len()-1)
			} else {
				return 0, &NameError{Name: fullName, Reason: "name contains a forbidden character"}
			}
		case 5:
			if *flags&nameRefspecPattern == 0 {
				if sanitized != nil {
					overwriteBuilderAt(sanitized, sanitized.Len()-1)
				} else {
					return 0, &NameError{Name: fullName, Reason: "name contains '*'"}
				}
			}

			*flags &^= nameRefspecPattern
		}

		last = ch
	}

out:
	componentLen := strings.IndexByte(name, '/')

	if componentLen < 0 {
		componentLen = len(name)
	}

	if componentLen == 0 {
		return 0, nil
	}

	if name[0] == '.' {
		if sanitized != nil {
			overwriteBuilderAt(sanitized, componentStart)
		} else {
			return 0, &NameError{Name: fullName, Reason: "component starts with '.'"}
		}
	}

	if componentLen >= len(lockSuffix) && name[componentLen-len(lockSuffix):componentLen] == lockSuffix {
		if sanitized == nil {
			return 0, &NameError{Name: fullName, Reason: "component ends with .lock"}
		}

		for strings.HasSuffix(sanitized.String(), lockSuffix) {
			truncateBuilder(sanitized, sanitized.Len()-len(lockSuffix))
		}
	}

	return componentLen, nil
}
