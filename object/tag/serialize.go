package tag

import (
	"bytes"
	"errors"
	"fmt"

	objectheader "codeberg.org/lindenii/furgit/object/header"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// SerializeWithoutHeader renders the raw tag body bytes.
func (tag *Tag) SerializeWithoutHeader() ([]byte, error) {
	if tag.Target.Size() == 0 {
		return nil, errors.New("object: tag: missing target id")
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "object %s\n", tag.Target.String())

	tyName, ok := objecttype.Name(tag.TargetType)
	if !ok {
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

// SerializeWithHeader renders the raw object (header + body).
func (tag *Tag) SerializeWithHeader() ([]byte, error) {
	body, err := tag.SerializeWithoutHeader()
	if err != nil {
		return nil, err
	}

	header, ok := objectheader.Encode(objecttype.TypeTag, int64(len(body)))
	if !ok {
		return nil, errors.New("object: tag: failed to encode object header")
	}

	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)

	return raw, nil
}
