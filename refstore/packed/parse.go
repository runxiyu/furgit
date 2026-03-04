package packed

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/ref"
)

// parsePackedRefs parses packed-refs content into detached refs.
func parsePackedRefs(r io.Reader, algo objectid.Algorithm) (map[string]ref.Detached, []ref.Detached, error) {
	byName := make(map[string]ref.Detached)
	ordered := make([]ref.Detached, 0, 32)

	br := bufio.NewReader(r)
	prev := -1
	lineNum := 0

	for {
		line, err := br.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, nil, err
		}

		if line == "" && err == io.EOF {
			break
		}

		lineNum++

		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")

		line = strings.TrimSpace(line)
		if line == "" {
			if err == io.EOF {
				break
			}

			continue
		}

		if strings.HasPrefix(line, "#") {
			if err == io.EOF {
				break
			}

			continue
		}

		if strings.HasPrefix(line, "^") {
			if prev < 0 {
				return nil, nil, fmt.Errorf("refstore/packed: line %d: peeled line without preceding ref", lineNum)
			}

			peeledHex := strings.TrimSpace(strings.TrimPrefix(line, "^"))

			peeled, parseErr := objectid.ParseHex(algo, peeledHex)
			if parseErr != nil {
				return nil, nil, fmt.Errorf("refstore/packed: line %d: invalid peeled oid: %w", lineNum, parseErr)
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

		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, nil, fmt.Errorf("refstore/packed: line %d: malformed entry", lineNum)
		}

		id, parseErr := objectid.ParseHex(algo, fields[0])
		if parseErr != nil {
			return nil, nil, fmt.Errorf("refstore/packed: line %d: invalid oid: %w", lineNum, parseErr)
		}

		name := fields[1]
		if name == "" {
			return nil, nil, fmt.Errorf("refstore/packed: line %d: empty ref name", lineNum)
		}

		if _, exists := byName[name]; exists {
			return nil, nil, fmt.Errorf("refstore/packed: line %d: duplicate ref %q", lineNum, name)
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
