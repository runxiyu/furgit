package furgit

import (
	"bytes"
	"errors"
	"fmt"
)

// Commit represents a Git commit object.
type Commit struct {
	// Tree represents the tree hash referenced by the commit.
	Tree Hash
	// Parents represents the parent commit hashes.
	// Commits that have 0 parents are root commits.
	// Commits that have >= 2 parents are merge commits.
	Parents []Hash
	// Author represents the author of the commit.
	Author Ident
	// Committer represents the committer of the commit.
	Committer Ident
	// Message represents the commit message.
	Message []byte
	// ExtraHeaders holds any extra headers present in the commit.
	ExtraHeaders []ExtraHeader
}

// StoredCommit represents a commit stored in the object database.
type StoredCommit struct {
	Commit
	hash Hash
}

// Hash returns the hash of the stored commit.
func (sCommit *StoredCommit) Hash() Hash {
	return sCommit.hash
}

// ObjectType returns the object type of the commit.
//
// It always returns ObjectTypeCommit.
func (commit *Commit) ObjectType() ObjectType {
	_ = commit
	return ObjectTypeCommit
}

func parseCommit(id Hash, body []byte, repo *Repository) (*StoredCommit, error) {
	c := new(StoredCommit)
	c.hash = id
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

func (commit *Commit) serialize() ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "tree %s\n", commit.Tree.String())
	for _, p := range commit.Parents {
		fmt.Fprintf(&buf, "parent %s\n", p.String())
	}
	buf.WriteString("author ")
	ab, err := commit.Author.Serialize()
	if err != nil {
		return nil, err
	}
	buf.Write(ab)
	buf.WriteByte('\n')
	buf.WriteString("committer ")
	cb, err := commit.Committer.Serialize()
	if err != nil {
		return nil, err
	}
	buf.Write(cb)
	buf.WriteByte('\n')
	buf.WriteByte('\n')
	buf.Write(commit.Message)

	return buf.Bytes(), nil
}

// Serialize renders the commit into its raw byte representation,
// including the header (i.e., "type size\0").
func (commit *Commit) Serialize() ([]byte, error) {
	body, err := commit.serialize()
	if err != nil {
		return nil, err
	}
	header, err := headerForType(ObjectTypeCommit, body)
	if err != nil {
		return nil, err
	}
	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)
	return raw, nil
}
