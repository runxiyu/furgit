package furgit

import (
	"bytes"
	"errors"
	"fmt"
)

// Commit mirrors the structure of a Git commit object.
type Commit struct {
	Hash         Hash
	Tree         Hash
	Parents      []Hash
	Author       Ident
	Committer    Ident
	Message      []byte
	ExtraHeaders []ExtraHeader
}

// ObjectType allows Commit to satisfy the Object interface.
func (commit *Commit) ObjectType() ObjectType {
	_ = commit
	return ObjCommit
}

func parseCommit(id Hash, body []byte, repo *Repository) (*Commit, error) {
	c := new(Commit)
	c.Hash = id
	i := 0
	for i < len(body) {
		rel := bytes.IndexByte(body[i:], '\n')
		if rel < 0 {
			return nil, errors.New("furgit: commit: missing newline")
		}
		line := body[i : i+rel]
		i += rel + 1
		if len(line) == 0 {
			break
		}

		switch {
		case bytes.HasPrefix(line, []byte("tree ")):
			treeID, err := repo.ParseHash(string(line[5:]))
			if err != nil {
				return nil, fmt.Errorf("furgit: commit: tree: %w", err)
			}
			c.Tree = treeID
		case bytes.HasPrefix(line, []byte("parent ")):
			parent, err := repo.ParseHash(string(line[7:]))
			if err != nil {
				return nil, fmt.Errorf("furgit: commit: parent: %w", err)
			}
			c.Parents = append(c.Parents, parent)
		case bytes.HasPrefix(line, []byte("author ")):
			idt, err := parseIdent(line[7:])
			if err != nil {
				return nil, fmt.Errorf("furgit: commit: author: %w", err)
			}
			c.Author = *idt
		case bytes.HasPrefix(line, []byte("committer ")):
			idt, err := parseIdent(line[10:])
			if err != nil {
				return nil, fmt.Errorf("furgit: commit: committer: %w", err)
			}
			c.Committer = *idt
		case bytes.HasPrefix(line, []byte("gpgsig ")), bytes.HasPrefix(line, []byte("gpgsig-sha256 ")):
			// TODO: handle this
			for i < len(body) {
				nextRel := bytes.IndexByte(body[i:], '\n')
				if nextRel < 0 {
					return nil, errors.New("furgit: commit: unterminated gpgsig")
				}
				if body[i] != ' ' {
					break
				}
				i += nextRel + 1
			}
		default:
			key, value, found := bytes.Cut(line, []byte{' '})
			if !found {
				return nil, errors.New("furgit: commit: malformed header")
			}
			c.ExtraHeaders = append(c.ExtraHeaders, ExtraHeader{Key: string(key), Value: value})
		}
	}

	if i > len(body) {
		return nil, ErrInvalidObject
	}

	c.Message = append([]byte(nil), body[i:]...)
	return c, nil
}

func commitBody(c *Commit) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "tree %s\n", c.Tree.String())
	for _, p := range c.Parents {
		fmt.Fprintf(&buf, "parent %s\n", p.String())
	}
	buf.WriteString("author ")
	ab, err := c.Author.Serialize()
	if err != nil {
		return nil, err
	}
	buf.Write(ab)
	buf.WriteByte('\n')
	buf.WriteString("committer ")
	cb, err := c.Committer.Serialize()
	if err != nil {
		return nil, err
	}
	buf.Write(cb)
	buf.WriteByte('\n')
	buf.WriteByte('\n')
	buf.Write(c.Message)

	return buf.Bytes(), nil
}

// Serialize renders a Commit into canonical Git format.
func (commit *Commit) Serialize() ([]byte, error) {
	body, err := commitBody(commit)
	if err != nil {
		return nil, err
	}
	header, err := headerForType(ObjCommit, body)
	if err != nil {
		return nil, err
	}
	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)
	return raw, nil
}
