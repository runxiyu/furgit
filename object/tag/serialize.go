package tag

import (
	"bytes"
	"errors"
	"fmt"

	objectheader "codeberg.org/lindenii/furgit/object/header"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// BytesWithoutHeader renders the raw tag body bytes.
func (tag *Tag) BytesWithoutHeader() ([]byte, error) {
	if tag.TargetID.Algorithm().Size() == 0 {
		return nil, errors.New("object: tag: missing target id")
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "object %s\n", tag.TargetID.String())

	tyName, ok := tag.TargetType.Name()
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

// BytesWithHeader renders the raw object (header + body).
func (tag *Tag) BytesWithHeader() ([]byte, error) {
	body, err := tag.BytesWithoutHeader()
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
