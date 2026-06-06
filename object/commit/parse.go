package commit

import (
	"bytes"
	"errors"
	"fmt"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/signature"
)

// ErrInvalidCommit indicates an attempt to parse an invalid commit.
var ErrInvalidCommit = errors.New("object/commit: invalid commit")

// Parse decodes a commit object body.
func Parse(body []byte, objectFormat id.ObjectFormat) (*Commit, error) {
	c := new(Commit)

	i := 0
	for i < len(body) {
		lineStart := i

		rel := bytes.IndexByte(body[i:], '\n')
		if rel < 0 {
			return nil, fmt.Errorf("%w: unterminated header line at offset %d", ErrInvalidCommit, lineStart)
		}

		line := body[i : i+rel]
		i += rel + 1

		if len(line) == 0 {
			break
		}

		key, value, found := bytes.Cut(line, []byte{' '})
		if !found {
			return nil, fmt.Errorf("%w: header line at offset %d has no ' ' separator", ErrInvalidCommit, lineStart)
		}

		switch string(key) {
		case "tree":
			id, err := objectFormat.FromString(string(value))
			if err != nil {
				return nil, fmt.Errorf("object/commit: tree: %w", err)
			}

			c.Tree = id
		case "parent":
			id, err := objectFormat.FromString(string(value))
			if err != nil {
				return nil, fmt.Errorf("object/commit: parent: %w", err)
			}

			c.Parents = append(c.Parents, id)
		case "author":
			idt, err := signature.Parse(value)
			if err != nil {
				return nil, fmt.Errorf("object/commit: author: %w", err)
			}

			c.Author = *idt
		case "committer":
			idt, err := signature.Parse(value)
			if err != nil {
				return nil, fmt.Errorf("object/commit: committer: %w", err)
			}

			c.Committer = *idt
		case "change-id":
			c.ChangeID = string(value)
		case "gpgsig", "gpgsig-sha256":
			for i < len(body) {
				nextRel := bytes.IndexByte(body[i:], '\n')
				if nextRel < 0 {
					return nil, fmt.Errorf("%w: unterminated signature header at offset %d", ErrInvalidCommit, i)
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
		return nil, fmt.Errorf("%w: header section extends past end of body", ErrInvalidCommit)
	}

	c.Message = append([]byte(nil), body[i:]...)

	return c, nil
}
