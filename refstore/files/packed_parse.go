package files

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/ref"
)

func parsePackedRefs(r io.Reader, algo objectid.Algorithm) (map[string]ref.Detached, []ref.Detached, error) {
	byName := make(map[string]ref.Detached)
	ordered := make([]ref.Detached, 0, 32)

	br := bufio.NewReader(r)
	prev := -1
	lineNum := 0
	hexsz := algo.Size() * 2

	for {
		line, err := br.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, nil, err
		}

		if line == "" && err == io.EOF {
			break
		}

		lineNum++
		hadNewline := strings.HasSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\n")

		if err == io.EOF && !hadNewline {
			return nil, nil, fmt.Errorf("refstore/files: line %d: unterminated line", lineNum)
		}

		if line == "" || strings.HasPrefix(line, "#") {
			if err == io.EOF {
				break
			}

			continue
		}

		if strings.HasPrefix(line, "^") {
			if prev < 0 {
				return nil, nil, fmt.Errorf("refstore/files: line %d: peeled line without preceding ref", lineNum)
			}

			if len(line) != hexsz+1 {
				return nil, nil, fmt.Errorf("refstore/files: line %d: malformed peeled line", lineNum)
			}

			peeled, parseErr := objectid.ParseHex(algo, line[1:])
			if parseErr != nil {
				return nil, nil, fmt.Errorf("refstore/files: line %d: invalid peeled oid: %w", lineNum, parseErr)
			}

			peeledCopy := peeled
			cur := ordered[prev]
			cur.Peeled = &peeledCopy
			ordered[prev] = cur
			byName[cur.Name()] = cur

			if err == io.EOF {
				break
			}

			continue
		}

		if len(line) < hexsz+2 {
			return nil, nil, fmt.Errorf("refstore/files: line %d: malformed entry", lineNum)
		}

		if line[hexsz] != ' ' {
			return nil, nil, fmt.Errorf("refstore/files: line %d: malformed entry", lineNum)
		}

		idText := line[:hexsz]

		name := line[hexsz+1:]
		if name == "" {
			return nil, nil, fmt.Errorf("refstore/files: line %d: empty ref name", lineNum)
		}

		id, parseErr := objectid.ParseHex(algo, idText)
		if parseErr != nil {
			return nil, nil, fmt.Errorf("refstore/files: line %d: invalid oid: %w", lineNum, parseErr)
		}

		if _, exists := byName[name]; exists {
			return nil, nil, fmt.Errorf("refstore/files: line %d: duplicate ref %q", lineNum, name)
		}

		detached := ref.Detached{
			RefName: name,
			ID:      id,
		}
		ordered = append(ordered, detached)
		prev = len(ordered) - 1
		byName[name] = detached

		if err == io.EOF {
			break
		}
	}

	return byName, ordered, nil
}
