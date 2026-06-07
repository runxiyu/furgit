package tag

import (
	"bytes"
	"errors"
	"fmt"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/signature"
	"lindenii.org/go/furgit/object/typ"
	refname "lindenii.org/go/furgit/ref/name"
)

// ErrInvalidTag indicates a malformed tag object.
var ErrInvalidTag = errors.New("object/tag: invalid tag")

// Parse decodes a tag object body.
func Parse(body []byte, objectFormat id.ObjectFormat) (*Tag, error) {
	t := new(Tag)

	i := 0

	var err error

	line, next, err := requiredHeaderLine(body, i, "object")
	if err != nil {
		return nil, err
	}

	t.TargetID, err = objectFormat.FromString(string(line))
	if err != nil {
		return nil, fmt.Errorf("%w: object: %w", ErrInvalidTag, err)
	}

	i = next

	line, next, err = requiredHeaderLine(body, i, "type")
	if err != nil {
		return nil, err
	}

	t.TargetType, err = typ.Parse(string(line))
	if err != nil {
		return nil, fmt.Errorf("%w: type: %w", ErrInvalidTag, err)
	}

	i = next

	line, next, err = requiredHeaderLine(body, i, "tag")
	if err != nil {
		return nil, err
	}

	_, err = refname.Tag(string(line))
	if err != nil {
		return nil, fmt.Errorf("%w: tag name: %w", ErrInvalidTag, err)
	}

	t.Name = append([]byte(nil), line...)
	i = next

	line, next, err = requiredHeaderLine(body, i, "tagger")
	if err != nil {
		return nil, err
	}

	tagger, err := signature.Parse(line)
	if err != nil {
		return nil, fmt.Errorf("%w: tagger: %w", ErrInvalidTag, err)
	}

	t.Tagger = *tagger
	i = next

	for i < len(body) {
		lineStart := i

		rel := bytes.IndexByte(body[i:], '\n')
		if rel < 0 {
			return nil, fmt.Errorf("%w: unterminated header line at offset %d", ErrInvalidTag, lineStart)
		}

		line := body[i : i+rel]
		i += rel + 1

		if len(line) == 0 {
			t.Message = append([]byte(nil), body[i:]...)

			return t, nil
		}

		key, value, found := bytes.Cut(line, []byte{' '})
		if !found {
			return nil, fmt.Errorf("%w: header line at offset %d has no ' ' separator", ErrInvalidTag, lineStart)
		}

		switch string(key) {
		case "object", "type", "tag", "tagger":
			return nil, fmt.Errorf("%w: unexpected %s header at offset %d", ErrInvalidTag, key, lineStart)
		case "gpgsig", "gpgsig-sha256":
			for i < len(body) {
				nextRel := bytes.IndexByte(body[i:], '\n')
				if nextRel < 0 {
					return nil, fmt.Errorf("%w: unterminated signature header at offset %d", ErrInvalidTag, i)
				}

				if body[i] != ' ' {
					break
				}

				i += nextRel + 1
			}
		default:
			t.ExtraHeaders = append(t.ExtraHeaders, ExtraHeader{
				Key:   string(key),
				Value: append([]byte(nil), value...),
			})
		}
	}

	return t, nil
}

func requiredHeaderLine(body []byte, offset int, want string) ([]byte, int, error) {
	rel := bytes.IndexByte(body[offset:], '\n')
	if rel < 0 {
		return nil, offset, fmt.Errorf("%w: unterminated %s header at offset %d", ErrInvalidTag, want, offset)
	}

	line := body[offset : offset+rel]
	next := offset + rel + 1

	key, value, found := bytes.Cut(line, []byte{' '})
	if !found {
		return nil, offset, fmt.Errorf("%w: %s header at offset %d has no ' ' separator", ErrInvalidTag, want, offset)
	}

	if string(key) != want {
		return nil, offset, fmt.Errorf("%w: expected %s header at offset %d, got %s", ErrInvalidTag, want, offset, key)
	}

	return value, next, nil
}
