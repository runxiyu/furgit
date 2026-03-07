package refname

import "strings"

func checkRefnameComponent(name string, flags *int, sanitized *strings.Builder, fullName string) (int, error) {
	var last byte

	componentStart := sanitizedLen(sanitized)

	for i := range len(name) {
		ch := name[i]
		disp := refnameDisposition(ch)

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
					overwriteLastByte(sanitized, '-')
				} else {
					return 0, &NameError{Name: fullName, Reason: "name contains '@{'"}
				}
			}
		case 4:
			if sanitized != nil {
				overwriteLastByte(sanitized, '-')
			} else {
				return 0, &NameError{Name: fullName, Reason: "name contains one forbidden character"}
			}
		case 5:
			if *flags&refnameRefspecPattern == 0 {
				if sanitized != nil {
					overwriteLastByte(sanitized, '-')
				} else {
					return 0, &NameError{Name: fullName, Reason: "name contains '*'"}
				}
			}

			*flags &^= refnameRefspecPattern
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
			overwriteBuilderAt(sanitized, componentStart, '-')
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
