package tag

import (
	"bytes"
	"slices"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/signed"
)

var signatureBeginLines = [][]byte{ //nolint:gochecknoglobals
	[]byte("-----BEGIN PGP SIGNATURE-----"),
	[]byte("-----BEGIN PGP MESSAGE-----"),
	[]byte("-----BEGIN SSH SIGNATURE-----"),
	[]byte("-----BEGIN SIGNED MESSAGE-----"),
}

// Parse parses one raw tag object body for signature extraction.
//
// Git stores the signature for storageFormat
// as an in-body ASCII-armored trailer,
// and may store additional signatures for other object formats
// in gpgsig* headers.
//
// The returned Tag remains valid only while body remains unchanged.
//
// Labels: Deps-Borrowed, Life-Parent.
func Parse(body []byte, storageFormat id.ObjectFormat) (*Tag, error) {
	tag := &Tag{
		body:       body,
		signatures: make(map[id.ObjectFormat][]byteRange),
	}

	signatureStart := len(body)
	for i := 0; i < len(body); {
		lineStart := i
		rel := bytes.IndexByte(body[i:], '\n')
		next := len(body)

		lineEnd := len(body)
		if rel >= 0 {
			lineEnd = i + rel
			next = lineEnd + 1
		}

		line := body[lineStart:lineEnd]
		if slices.ContainsFunc(signatureBeginLines, func(begin []byte) bool {
			return bytes.HasPrefix(line, begin)
		}) {
			signatureStart = lineStart
		}

		i = next
	}

	payloadStart := 0

	payloadEnd := signatureStart
	if signatureStart == len(body) {
		payloadEnd = len(body)
	}

	for i := 0; i < payloadEnd; {
		lineStart := i
		rel := bytes.IndexByte(body[i:payloadEnd], '\n')
		next := payloadEnd

		lineEnd := payloadEnd
		if rel >= 0 {
			lineEnd = i + rel
			next = lineEnd + 1
		}

		line := body[lineStart:lineEnd]
		i = next

		if len(line) == 0 {
			break
		}

		if line[0] == ' ' {
			continue
		}

		key, valueStart, found := bytes.Cut(line, []byte{' '})
		if !found {
			continue
		}

		objectFormat, ok := signed.ParseSignatureHeaderName(string(key))
		if !ok {
			continue
		}

		tag.appendPayloadRange(payloadStart, lineStart)
		tag.signatures[objectFormat] = append(tag.signatures[objectFormat], byteRange{
			start: lineEnd - len(valueStart),
			end:   next,
		})

		for i < payloadEnd {
			rel := bytes.IndexByte(body[i:payloadEnd], '\n')
			next = payloadEnd

			lineEnd = payloadEnd
			if rel >= 0 {
				lineEnd = i + rel
				next = lineEnd + 1
			}

			cont := body[i:lineEnd]
			if len(cont) == 0 || cont[0] != ' ' {
				break
			}

			tag.signatures[objectFormat] = append(tag.signatures[objectFormat], byteRange{
				start: i + 1,
				end:   next,
			})

			i = next
		}

		payloadStart = i
	}

	tag.appendPayloadRange(payloadStart, payloadEnd)

	if signatureStart != len(body) {
		tag.signatures[storageFormat] = append(tag.signatures[storageFormat], byteRange{
			start: signatureStart,
			end:   len(body),
		})
	}

	return tag, nil
}

func (tag *Tag) appendPayloadRange(start, end int) {
	if start >= end {
		return
	}

	tag.payload = append(tag.payload, byteRange{start: start, end: end})
}
