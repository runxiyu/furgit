package object

import (
	"bytes"
	"errors"
	"fmt"

	"codeberg.org/lindenii/furgit/objectheader"
	"codeberg.org/lindenii/furgit/objecttype"
)

// SerializeWithoutHeader renders the raw commit body bytes.
func (commit *Commit) SerializeWithoutHeader() ([]byte, error) {
	var buf bytes.Buffer

	if commit.Tree.Size() == 0 {
		return nil, errors.New("object: commit: missing tree id")
	}

	fmt.Fprintf(&buf, "tree %s\n", commit.Tree.String())

	for _, parent := range commit.Parents {
		fmt.Fprintf(&buf, "parent %s\n", parent.String())
	}

	authorBytes, err := commit.Author.Serialize()
	if err != nil {
		return nil, err
	}

	buf.WriteString("author ")
	buf.Write(authorBytes)
	buf.WriteByte('\n')

	committerBytes, err := commit.Committer.Serialize()
	if err != nil {
		return nil, err
	}

	buf.WriteString("committer ")
	buf.Write(committerBytes)
	buf.WriteByte('\n')

	if commit.ChangeID != "" {
		buf.WriteString("change-id ")
		buf.WriteString(commit.ChangeID)
		buf.WriteByte('\n')
	}

	for _, h := range commit.ExtraHeaders {
		if h.Key == "" {
			return nil, errors.New("object: commit: extra header has empty key")
		}

		buf.WriteString(h.Key)
		buf.WriteByte(' ')
		buf.Write(h.Value)
		buf.WriteByte('\n')
	}

	buf.WriteByte('\n')
	buf.Write(commit.Message)

	return buf.Bytes(), nil
}

// SerializeWithHeader renders the raw object (header + body).
func (commit *Commit) SerializeWithHeader() ([]byte, error) {
	body, err := commit.SerializeWithoutHeader()
	if err != nil {
		return nil, err
	}

	header, ok := objectheader.Encode(objecttype.TypeCommit, int64(len(body)))
	if !ok {
		return nil, errors.New("object: commit: failed to encode object header")
	}

	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)

	return raw, nil
}
