package main

import (
	"fmt"
	"os"
	"strings"

	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/object/blob"
	"codeberg.org/lindenii/furgit/object/commit"
	"codeberg.org/lindenii/furgit/object/stored"
	"codeberg.org/lindenii/furgit/object/tag"
	"codeberg.org/lindenii/furgit/object/tree"
)

func printStored(s *stored.Stored[object.Object]) {
	var b strings.Builder

	id := s.ID()
	ty := s.Object().ObjectType()

	tyName, ok := ty.Name()
	if !ok {
		tyName = fmt.Sprintf("type %d", ty)
	}

	fmt.Fprintf(&b, "id: %s\n", id)
	fmt.Fprintf(&b, "type: %s\n", tyName)

	switch obj := s.Object().(type) {
	case *blob.Blob:
		blob := obj
		fmt.Fprintf(&b, "size: %d\n", len(blob.Data))
		fmt.Fprintf(&b, "data: %q\n", string(blob.Data))
	case *tree.Tree:
		tree := obj
		fmt.Fprintf(&b, "entries: %d\n", len(tree.Entries))

		for _, entry := range tree.Entries {
			fmt.Fprintf(&b, "%06o %s\t%s\n", entry.Mode, entry.ID, entry.Name)
		}
	case *commit.Commit:
		commit := obj
		fmt.Fprintf(&b, "tree: %s\n", commit.Tree)

		for _, parent := range commit.Parents {
			fmt.Fprintf(&b, "parent: %s\n", parent)
		}

		fmt.Fprintf(&b, "author: %s <%s>\n", commit.Author.Name, commit.Author.Email)
		fmt.Fprintf(&b, "committer: %s <%s>\n", commit.Committer.Name, commit.Committer.Email)
		fmt.Fprintf(&b, "message:\n%s\n", string(commit.Message))
	case *tag.Tag:
		tag := obj

		targetTy, ok := tag.TargetType.Name()
		if !ok {
			targetTy = fmt.Sprintf("type %d", tag.TargetType)
		}

		fmt.Fprintf(&b, "target: %s (%s)\n", tag.TargetID, targetTy)
		fmt.Fprintf(&b, "name: %s\n", tag.Name)

		if tag.Tagger != nil {
			fmt.Fprintf(&b, "tagger: %s <%s>\n", tag.Tagger.Name, tag.Tagger.Email)
		}

		fmt.Fprintf(&b, "message:\n%s\n", string(tag.Message))
	default:
		fmt.Fprintf(&b, "%#v\n", obj)
	}

	_, _ = os.Stdout.WriteString(b.String())
}
