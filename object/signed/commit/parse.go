package signedcommit

import (
	"bytes"

	objectid "codeberg.org/lindenii/furgit/object/id"
)

// Parse parses one raw commit object body for signature extraction.
//
// The returned Commit remains valid only while body remains unchanged.
//
// Labels: Deps-Borrowed, Life-Parent.
func Parse(body []byte) (*Commit, error) {
	commit := &Commit{
		body:       body,
		signatures: make(map[objectid.Algorithm][]byteRange),
	}

	payloadStart := 0
	i := 0

	for i < len(body) {
		lineStart := i

		rel := bytes.IndexByte(body[i:], '\n')
		next := len(body)
		lineEnd := len(body)
		if rel >= 0 {
			lineEnd = i + rel
			next = lineEnd + 1
		}

		line := body[lineStart:lineEnd]
		i = next

		if len(line) == 0 {
			commit.appendPayloadRange(payloadStart, len(body))

			return commit, nil
		}

		if line[0] == ' ' {
			continue
		}

		if !bytes.HasPrefix(line, []byte("gpgsig")) {
			continue
		}

		commit.appendPayloadRange(payloadStart, lineStart)

		key, valueStart, found := bytes.Cut(line, []byte{' '})
		if found {
			if algo, ok := objectid.ParseSignatureHeaderName(string(key)); ok {
				commit.signatures[algo] = append(commit.signatures[algo], byteRange{
					start: lineEnd - len(valueStart),
					end:   next,
				})
			}
		}

		for i < len(body) {
			rel := bytes.IndexByte(body[i:], '\n')
			next = len(body)
			lineEnd = len(body)
			if rel >= 0 {
				lineEnd = i + rel
				next = lineEnd + 1
			}

			contStart := i
			cont := body[contStart:lineEnd]
			if len(cont) == 0 || cont[0] != ' ' {
				break
			}

			if found {
				if algo, ok := objectid.ParseSignatureHeaderName(string(key)); ok {
					commit.signatures[algo] = append(commit.signatures[algo], byteRange{
						start: contStart + 1,
						end:   next,
					})
				}
			}

			i = next
		}

		payloadStart = i
	}

	commit.appendPayloadRange(payloadStart, len(body))

	return commit, nil
}

func (commit *Commit) appendPayloadRange(start, end int) {
	if start >= end {
		return
	}

	commit.payload = append(commit.payload, byteRange{start: start, end: end})
}
