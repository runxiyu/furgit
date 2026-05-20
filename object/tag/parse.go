package tag

import (
	"bytes"
	"errors"
	"fmt"

	objectid "codeberg.org/lindenii/furgit/object/id"
	objectsignature "codeberg.org/lindenii/furgit/object/signature"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// Parse decodes a tag object body.
func Parse(body []byte, algo objectid.Algorithm) (*Tag, error) {
	t := new(Tag)
	i := 0

	var haveTarget, haveType bool

	for i < len(body) {
		rel := bytes.IndexByte(body[i:], '\n')
		if rel < 0 {
			return nil, errors.New("object: tag: missing newline")
		}

		line := body[i : i+rel]
		i += rel + 1

		if len(line) == 0 {
			break
		}

		key, value, found := bytes.Cut(line, []byte{' '})
		if !found {
			return nil, errors.New("object: tag: malformed header")
		}

		switch string(key) {
		case "object":
			id, err := objectid.ParseHex(algo, string(value))
			if err != nil {
				return nil, fmt.Errorf("object: tag: object: %w", err)
			}

			t.TargetID = id
			haveTarget = true
		case "type":
			ty, ok := objecttype.Parse(string(value))
			if !ok {
				return nil, errors.New("object: tag: unknown target type")
			}

			t.TargetType = ty
			haveType = true
		case "tag":
			t.Name = append([]byte(nil), value...)
		case "tagger":
			idt, err := objectsignature.Parse(value)
			if err != nil {
				return nil, fmt.Errorf("object: tag: tagger: %w", err)
			}

			t.Tagger = idt
		case "gpgsig", "gpgsig-sha256":
			for i < len(body) {
				nextRel := bytes.IndexByte(body[i:], '\n')
				if nextRel < 0 {
					return nil, errors.New("object: tag: unterminated gpgsig")
				}

				if body[i] != ' ' {
					break
				}

				i += nextRel + 1
			}
		default:
			// Ignore unknown headers for now.
		}
	}

	if !haveTarget || !haveType {
		return nil, errors.New("object: tag: missing required headers")
	}

	t.Message = append([]byte(nil), body[i:]...)

	return t, nil
}
