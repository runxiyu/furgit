package alternates

import (
	"path/filepath"
	"strings"
)

// Parse returns the object directories named by an alternates file.
//
// Each line names one directory.
// A line opening with "#" is a comment,
// and a line opening with a quote is C-style quoted,
// which lets a name carry characters a bare line could not hold.
// A line whose quoting is broken is taken literally,
// as are all other lines.
//
// A relative name is joined to relativeBase,
// which is the objects directory holding the file.
// Names are returned as they were written,
// neither normalized nor checked against the filesystem.
func Parse(data []byte, relativeBase string) []string {
	paths := []string{}

	for line := range strings.SplitSeq(string(data), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		path := line

		if strings.HasPrefix(line, `"`) {
			unquoted, ok := unquote(line)
			if ok {
				path = unquoted
			}
		}

		if path == "" {
			continue
		}

		if !filepath.IsAbs(path) {
			path = filepath.Join(relativeBase, path)
		}

		paths = append(paths, path)
	}

	return paths
}

// unquote decodes one C-style quoted name,
// reporting whether it was well formed.
func unquote(quoted string) (string, bool) {
	if !strings.HasPrefix(quoted, `"`) {
		return "", false
	}

	out := []byte{}

	for i := 1; i < len(quoted); {
		ch := quoted[i]
		i++

		if ch == '"' {
			return string(out), true
		}

		if ch != '\\' {
			out = append(out, ch)

			continue
		}

		if i >= len(quoted) {
			return "", false
		}

		escape := quoted[i]
		i++

		decoded, width, ok := decodeEscape(escape, quoted[i:])
		if !ok {
			return "", false
		}

		out = append(out, decoded)
		i += width
	}

	return "", false
}

// decodeEscape decodes one escape following a backslash,
// reporting how many further bytes it consumed.
func decodeEscape(escape byte, rest string) (byte, int, bool) {
	switch escape {
	case 'a':
		return '\a', 0, true
	case 'b':
		return '\b', 0, true
	case 'f':
		return '\f', 0, true
	case 'n':
		return '\n', 0, true
	case 'r':
		return '\r', 0, true
	case 't':
		return '\t', 0, true
	case 'v':
		return '\v', 0, true
	case '\\', '"':
		return escape, 0, true
	case '0', '1', '2', '3':
		if len(rest) < 2 || !isOctal(rest[0]) || !isOctal(rest[1]) {
			return 0, 0, false
		}

		return (escape-'0')<<6 | (rest[0]-'0')<<3 | (rest[1] - '0'), 2, true
	default:
		return 0, 0, false
	}
}

func isOctal(ch byte) bool {
	return ch >= '0' && ch <= '7'
}
