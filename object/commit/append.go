package commit

import (
	"fmt"

	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/typ"
)

// AppendWithoutHeader renders the raw commit body bytes.
func (commit *Commit) AppendWithoutHeader(dst []byte) ([]byte, error) {
	dst = fmt.Appendf(dst, "tree %s\n", commit.Tree.String())

	for _, parent := range commit.Parents {
		dst = fmt.Appendf(dst, "parent %s\n", parent.String())
	}

	dst = append(dst, []byte("author ")...)

	dst, err := commit.Author.Append(dst)
	if err != nil {
		return dst, fmt.Errorf("object/commit: append author: %w", err)
	}

	dst = append(dst, byte('\n'))

	dst = append(dst, []byte("committer ")...)

	dst, err = commit.Committer.Append(dst)
	if err != nil {
		return dst, fmt.Errorf("object/commit: append committer: %w", err)
	}

	dst = append(dst, byte('\n'))

	if commit.ChangeID != "" {
		dst = append(dst, []byte("change-id ")...)
		dst = append(dst, commit.ChangeID...)
		dst = append(dst, byte('\n'))
	}

	for _, h := range commit.ExtraHeaders {
		// GIGO on empty keys and such.
		dst = append(dst, []byte(h.Key)...)
		dst = append(dst, byte(' '))
		dst = append(dst, h.Value...)
		dst = append(dst, byte('\n'))
	}

	dst = append(dst, byte('\n'))
	dst = append(dst, commit.Message...)

	return dst, nil
}

// AppendWithHeader renders the raw object (header + body).
func (commit *Commit) AppendWithHeader(dst []byte) ([]byte, error) {
	// TODO: Try to not allocate?
	body, err := commit.AppendWithoutHeader(nil)
	if err != nil {
		return dst, err
	}

	dst = header.Append(dst, typ.Commit, len(body))

	return append(dst, body...), nil
}
