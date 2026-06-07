package commit

import (
	"bytes"
	"errors"
	"fmt"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/signature"
)

// ErrInvalidCommit indicates a malformed commit object.
var ErrInvalidCommit = errors.New("object/commit: invalid commit")

// Parse decodes a commit object body.
func Parse(body []byte, objectFormat id.ObjectFormat) (*Commit, error) {
	c := new(Commit)

	i := 0
	state := parseStateTree
	sawHeaderEnd := false

	for i < len(body) {
		lineStart := i

		rel := bytes.IndexByte(body[i:], '\n')
		if rel < 0 {
			return nil, fmt.Errorf("%w: unterminated header line at offset %d", ErrInvalidCommit, lineStart)
		}

		line := body[i : i+rel]
		i += rel + 1

		if len(line) == 0 {
			sawHeaderEnd = true

			break
		}

		key, value, found := bytes.Cut(line, []byte{' '})
		if !found {
			return nil, fmt.Errorf("%w: header line at offset %d has no ' ' separator", ErrInvalidCommit, lineStart)
		}

		switch string(key) {
		case "tree":
			if state != parseStateTree {
				return nil, fmt.Errorf("%w: unexpected tree header at offset %d", ErrInvalidCommit, lineStart)
			}

			id, err := objectFormat.FromString(string(value))
			if err != nil {
				return nil, fmt.Errorf("%w: tree: %w", ErrInvalidCommit, err)
			}

			c.Tree = id
			state = parseStateParentOrAuthor
		case "parent":
			if state != parseStateParentOrAuthor {
				return nil, fmt.Errorf("%w: unexpected parent header at offset %d", ErrInvalidCommit, lineStart)
			}

			id, err := objectFormat.FromString(string(value))
			if err != nil {
				return nil, fmt.Errorf("%w: parent: %w", ErrInvalidCommit, err)
			}

			c.Parents = append(c.Parents, id)
		case "author":
			if state != parseStateParentOrAuthor {
				return nil, fmt.Errorf("%w: unexpected author header at offset %d", ErrInvalidCommit, lineStart)
			}

			idt, err := signature.Parse(value)
			if err != nil {
				return nil, fmt.Errorf("%w: author: %w", ErrInvalidCommit, err)
			}

			c.Author = *idt
			state = parseStateCommitter
		case "committer":
			if state != parseStateCommitter {
				return nil, fmt.Errorf("%w: unexpected committer header at offset %d", ErrInvalidCommit, lineStart)
			}

			idt, err := signature.Parse(value)
			if err != nil {
				return nil, fmt.Errorf("%w: committer: %w", ErrInvalidCommit, err)
			}

			c.Committer = *idt
			state = parseStateExtra
		case "change-id":
			if state != parseStateExtra {
				return nil, fmt.Errorf("%w: unexpected change-id header at offset %d", ErrInvalidCommit, lineStart)
			}

			c.ChangeID = string(value)
		case "gpgsig", "gpgsig-sha256":
			if state != parseStateExtra {
				return nil, fmt.Errorf("%w: unexpected %s header at offset %d", ErrInvalidCommit, key, lineStart)
			}

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
			if state != parseStateExtra {
				return nil, fmt.Errorf("%w: unexpected %s header at offset %d", ErrInvalidCommit, key, lineStart)
			}

			c.ExtraHeaders = append(c.ExtraHeaders, ExtraHeader{
				Key:   string(key),
				Value: append([]byte(nil), value...),
			})
		}
	}

	if !sawHeaderEnd {
		return nil, fmt.Errorf("%w: missing blank line before message", ErrInvalidCommit)
	}

	switch state {
	case parseStateTree:
		return nil, fmt.Errorf("%w: missing tree header", ErrInvalidCommit)
	case parseStateParentOrAuthor:
		return nil, fmt.Errorf("%w: missing author header", ErrInvalidCommit)
	case parseStateCommitter:
		return nil, fmt.Errorf("%w: missing committer header", ErrInvalidCommit)
	case parseStateExtra:
	default:
		panic("unreachable parse state")
	}

	c.Message = append([]byte(nil), body[i:]...)

	return c, nil
}

type parseState uint8

const (
	parseStateTree parseState = iota
	parseStateParentOrAuthor
	parseStateCommitter
	parseStateExtra
)
