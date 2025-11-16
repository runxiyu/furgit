package furgit

import (
	"bytes"
	"errors"
	"fmt"
)

// Tag represents a Git annotated tag object.
type Tag struct {
	// Target represents the hash of the object being tagged.
	Target Hash
	// TargetType represents the type of the object being tagged.
	TargetType ObjectType
	// Name represents the name of the tag.
	Name []byte
	// Tagger represents the identity of the tagger.
	Tagger *Ident
	// Message represents the tag message.
	Message []byte
}

// TODO: ExtraHeaders and signatures

// StoredTag represents a tag stored in the object database.
type StoredTag struct {
	Tag
	hash Hash
}

// Hash returns the hash of the stored tag.
func (sTag *StoredTag) Hash() Hash {
	return sTag.hash
}

// ObjectType returns the object type of the tag.
//
// It always returns ObjectTypeTag.
func (tag *Tag) ObjectType() ObjectType {
	_ = tag
	return ObjectTypeTag
}

// parseTag parses a tag object body.
func parseTag(id Hash, body []byte, repo *Repository) (*StoredTag, error) {
	t := new(StoredTag)
	t.hash = id
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
				t.TargetType = ObjectTypeCommit
			case "tree":
				t.TargetType = ObjectTypeTree
			case "blob":
				t.TargetType = ObjectTypeBlob
			case "tag":
				t.TargetType = ObjectTypeTag
			default:
				t.TargetType = ObjectTypeInvalid
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
	case ObjectTypeCommit:
		buf.WriteString("commit")
	case ObjectTypeTree:
		buf.WriteString("tree")
	case ObjectTypeBlob:
		buf.WriteString("blob")
	case ObjectTypeTag:
		buf.WriteString("tag")
	case ObjectTypeInvalid, ObjectTypeFuture, ObjectTypeOfsDelta, ObjectTypeRefDelta:
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
		tb, err := t.Tagger.Serialize()
		if err != nil {
			return nil, err
		}
		buf.Write(tb)
		buf.WriteByte('\n')
	}
	buf.WriteByte('\n')
	buf.Write(t.Message)

	return buf.Bytes(), nil
}

// Serialize renders the tag into its raw byte representation,
// including the header (i.e., "type size\0").
func (tag *Tag) Serialize() ([]byte, error) {
	body, err := tagBody(tag)
	if err != nil {
		return nil, err
	}
	header, err := headerForType(ObjectTypeTag, body)
	if err != nil {
		return nil, err
	}
	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)
	return raw, nil
}
