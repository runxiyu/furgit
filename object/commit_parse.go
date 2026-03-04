package object

import (
	"bytes"
	"errors"
	"fmt"

	"codeberg.org/lindenii/furgit/objectid"
)

// ParseCommit decodes a commit object body.
func ParseCommit(body []byte, algo objectid.Algorithm) (*Commit, error) {
	c := new(Commit)

	i := 0
	for i < len(body) {
		rel := bytes.IndexByte(body[i:], '\n')
		if rel < 0 {
			return nil, errors.New("object: commit: missing newline")
		}

		line := body[i : i+rel]
		i += rel + 1

		if len(line) == 0 {
			break
		}

		key, value, found := bytes.Cut(line, []byte{' '})
		if !found {
			return nil, errors.New("object: commit: malformed header")
		}

		switch string(key) {
		case "tree":
			id, err := objectid.ParseHex(algo, string(value))
			if err != nil {
				return nil, fmt.Errorf("object: commit: tree: %w", err)
			}

			c.Tree = id
		case "parent":
			id, err := objectid.ParseHex(algo, string(value))
			if err != nil {
				return nil, fmt.Errorf("object: commit: parent: %w", err)
			}

			c.Parents = append(c.Parents, id)
		case "author":
			idt, err := ParseSignature(value)
			if err != nil {
				return nil, fmt.Errorf("object: commit: author: %w", err)
			}

			c.Author = *idt
		case "committer":
			idt, err := ParseSignature(value)
			if err != nil {
				return nil, fmt.Errorf("object: commit: committer: %w", err)
			}

			c.Committer = *idt
		case "change-id":
			c.ChangeID = string(value)
		case "gpgsig", "gpgsig-sha256":
			for i < len(body) {
				nextRel := bytes.IndexByte(body[i:], '\n')
				if nextRel < 0 {
					return nil, errors.New("object: commit: unterminated gpgsig")
				}

				if body[i] != ' ' {
					break
				}

				i += nextRel + 1
			}
		default:
			c.ExtraHeaders = append(c.ExtraHeaders, ExtraHeader{
				Key:   string(key),
				Value: append([]byte(nil), value...),
			})
		}
	}

	if i > len(body) {
		return nil, errors.New("object: commit: parser position out of bounds")
	}

	c.Message = append([]byte(nil), body[i:]...)

	return c, nil
}
