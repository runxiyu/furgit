package object

import (
	"bytes"
	"fmt"
)

func (tag *Tag) serialize() ([]byte, error) {
	if tag.Target.Size() == 0 {
		return nil, ErrInvalidObject
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "object %s\n", tag.Target.String())

	tyName, err := typeName(tag.TargetType)
	if err != nil {
		return nil, fmt.Errorf("object: tag: invalid target type %d", tag.TargetType)
	}
	buf.WriteString("type ")
	buf.WriteString(tyName)
	buf.WriteByte('\n')

	buf.WriteString("tag ")
	buf.Write(tag.Name)
	buf.WriteByte('\n')

	if tag.Tagger != nil {
		taggerBytes, err := tag.Tagger.Serialize()
		if err != nil {
			return nil, err
		}
		buf.WriteString("tagger ")
		buf.Write(taggerBytes)
		buf.WriteByte('\n')
	}

	buf.WriteByte('\n')
	buf.Write(tag.Message)
	return buf.Bytes(), nil
}

// Serialize renders the raw object (header + body).
func (tag *Tag) Serialize() ([]byte, error) {
	body, err := tag.serialize()
	if err != nil {
		return nil, err
	}
	header, err := headerForType(TypeTag, body)
	if err != nil {
		return nil, err
	}
	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)
	return raw, nil
}
