package object

import (
	"bytes"
	"fmt"

	"codeberg.org/lindenii/furgit/objecttype"
)

func (commit *Commit) serialize() ([]byte, error) {
	var buf bytes.Buffer

	if commit.Tree.Size() == 0 {
		return nil, ErrInvalidObject
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
			return nil, ErrInvalidObject
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

// Serialize renders the raw object (header + body).
func (commit *Commit) Serialize() ([]byte, error) {
	body, err := commit.serialize()
	if err != nil {
		return nil, err
	}
	header, err := headerForType(objecttype.TypeCommit, body)
	if err != nil {
		return nil, err
	}
	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)
	return raw, nil
}
