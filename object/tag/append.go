package tag

import (
	"fmt"

	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/typ"
)

// AppendWithoutHeader renders the raw tag body bytes.
func (tag *Tag) AppendWithoutHeader(dst []byte) ([]byte, error) {
	dst = fmt.Appendf(dst, "object %s\n", tag.TargetID.String())
	dst = append(dst, []byte("type ")...)
	dst = append(dst, tag.TargetType.Name()...)
	dst = append(dst, byte('\n'))
	dst = append(dst, []byte("tag ")...)
	dst = append(dst, tag.Name...)
	dst = append(dst, byte('\n'))
	dst = append(dst, []byte("tagger ")...)

	dst, err := tag.Tagger.Append(dst)
	if err != nil {
		return dst, fmt.Errorf("object/tag: append tagger: %w", err)
	}

	dst = append(dst, byte('\n'))

	for _, h := range tag.ExtraHeaders {
		// GIGO on empty keys and such.
		dst = append(dst, []byte(h.Key)...)
		dst = append(dst, byte(' '))
		dst = append(dst, h.Value...)
		dst = append(dst, byte('\n'))
	}

	dst = append(dst, byte('\n'))
	dst = append(dst, tag.Message...)

	return dst, nil
}

// AppendWithHeader renders the raw object (header + body).
func (tag *Tag) AppendWithHeader(dst []byte) ([]byte, error) {
	// TODO: Try to not allocate?
	body, err := tag.AppendWithoutHeader(nil)
	if err != nil {
		return dst, err
	}

	dst = header.Append(dst, typ.Tag, len(body))

	return append(dst, body...), nil
}
