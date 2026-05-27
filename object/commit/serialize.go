package commit

import (
	"bytes"
	"errors"
	"fmt"

	objectheader "lindenii.org/go/furgit/object/header"
	objecttype "lindenii.org/go/furgit/object/type"
)

// BytesWithoutHeader renders the raw commit body bytes.
func (commit *Commit) BytesWithoutHeader() ([]byte, error) {
	var buf bytes.Buffer

	if commit.Tree.Algorithm().Size() == 0 {
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

// BytesWithHeader renders the raw object (header + body).
func (commit *Commit) BytesWithHeader() ([]byte, error) {
	body, err := commit.BytesWithoutHeader()
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
