package furgit

import (
	"bytes"
	"errors"
	"fmt"
)

// Tag models an annotated Git tag object.
type Tag struct {
	Hash       Hash
	Target     Hash
	TargetType ObjType
	Name       []byte
	Tagger     *Ident
	Message    []byte
}

// ObjType allows Tag to satisfy the Object interface.
func (*Tag) ObjType() ObjType {
	return ObjTag
}

// parseTag parses a tag object body.
func parseTag(id Hash, body []byte, repo *Repository) (*Tag, error) {
	t := new(Tag)
	t.Hash = id
	i := 0
	var haveTarget, haveType bool

	for i < len(body) {
		rel := bytes.IndexByte(body[i:], '\n')
		if rel < 0 {
			return nil, errors.New("furgit: tag: missing newline")
		}
		line := body[i : i+rel]
		i += rel + 1
		if len(line) == 0 {
			break
		}

		switch {
		case bytes.HasPrefix(line, []byte("object ")):
			hash, err := repo.ParseHash(string(line[7:]))
			if err != nil {
				return nil, fmt.Errorf("furgit: tag: object: %w", err)
			}
			t.Target = hash
			haveTarget = true
		case bytes.HasPrefix(line, []byte("type ")):
			switch string(line[5:]) {
			case "commit":
				t.TargetType = ObjCommit
			case "tree":
				t.TargetType = ObjTree
			case "blob":
				t.TargetType = ObjBlob
			case "tag":
				t.TargetType = ObjTag
			default:
				t.TargetType = ObjInvalid
				return nil, errors.New("furgit: tag: unknown target type")
			}
			haveType = true
		case bytes.HasPrefix(line, []byte("tag ")):
			t.Name = append([]byte(nil), line[4:]...)
		case bytes.HasPrefix(line, []byte("tagger ")):
			idt, err := parseIdent(line[7:])
			if err != nil {
				return nil, fmt.Errorf("furgit: tag: tagger: %w", err)
			}
			t.Tagger = idt
		case bytes.HasPrefix(line, []byte("gpgsig ")), bytes.HasPrefix(line, []byte("gpgsig-sha256 ")):
			for i < len(body) {
				nextRel := bytes.IndexByte(body[i:], '\n')
				if nextRel < 0 {
					return nil, errors.New("furgit: tag: unterminated gpgsig")
				}
				if body[i] != ' ' {
					break
				}
				i += nextRel + 1
			}
		default:
			// ignore unknown headers
		}
	}

	if !haveTarget || !haveType {
		return nil, errors.New("furgit: tag: missing required headers")
	}

	t.Message = append([]byte(nil), body[i:]...)
	return t, nil
}

func tagBody(t *Tag) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "object %s\n", t.Target.String())
	buf.WriteString("type ")
	switch t.TargetType {
	case ObjCommit:
		buf.WriteString("commit")
	case ObjTree:
		buf.WriteString("tree")
	case ObjBlob:
		buf.WriteString("blob")
	case ObjTag:
		buf.WriteString("tag")
	case ObjInvalid, ObjFuture, ObjOfsDelta, ObjRefDelta:
		return nil, fmt.Errorf("furgit: tag: invalid target type %d", t.TargetType)
	default:
		return nil, fmt.Errorf("furgit: tag: invalid target type %d", t.TargetType)
	}
	buf.WriteByte('\n')
	buf.WriteString("tag ")
	buf.Write(t.Name)
	buf.WriteByte('\n')
	if t.Tagger != nil {
		buf.WriteString("tagger ")
		buf.Write(t.Tagger.Serialize())
		buf.WriteByte('\n')
	}
	buf.WriteByte('\n')
	buf.Write(t.Message)

	return buf.Bytes(), nil
}

// Serialize renders a Tag into canonical Git format.
func (tag *Tag) Serialize() ([]byte, error) {
	body, err := tagBody(tag)
	if err != nil {
		return nil, err
	}
	header, err := headerForType(ObjTag, body)
	if err != nil {
		return nil, err
	}
	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)
	return raw, nil
}
