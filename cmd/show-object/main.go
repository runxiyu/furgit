package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstored"
	"codeberg.org/lindenii/furgit/objecttype"
	"codeberg.org/lindenii/furgit/repository"
)

func main() {
	repoPath := flag.String("r", "", "path to git dir (.git or bare repo root)")
	name := flag.String("h", "", "reference name or object id")
	flag.Parse()

	if *repoPath == "" || *name == "" {
		log.Fatal("must provide -r <repo> and -h <ref-or-object-id>")
	}

	if err := run(repoPath, name); err != nil {
		log.Fatalf("run: %v", err)
	}
}

func run(repoPath, name *string) error {
	root, err := os.OpenRoot(*repoPath)
	if err != nil {
		return fmt.Errorf("open repo root: %w", err)
	}
	defer func() { _ = root.Close() }()

	repo, err := repository.Open(root)
	if err != nil {
		return fmt.Errorf("open repository: %w", err)
	}

	id, err := resolveInput(repo, *name)
	if err != nil {
		_ = repo.Close()
		return fmt.Errorf("resolve %q: %w", *name, err)
	}

	stored, err := repo.ReadStored(id)
	if err != nil {
		_ = repo.Close()
		return fmt.Errorf("read object %s: %w", id, err)
	}

	printStored(stored)
	if err := repo.Close(); err != nil {
		return fmt.Errorf("close repository: %w", err)
	}

	return nil
}

func resolveInput(repo *repository.Repository, input string) (objectid.ObjectID, error) {
	if id, err := objectid.ParseHex(repo.Algorithm(), strings.TrimSpace(input)); err == nil {
		return id, nil
	}
	resolved, err := repo.Refs().ResolveFully(input)
	if err != nil {
		return objectid.ObjectID{}, err
	}
	return resolved.ID, nil
}

func printStored(stored objectstored.StoredObject) {
	var b strings.Builder

	id := stored.ID()
	ty := stored.Object().ObjectType()
	tyName, ok := objecttype.Name(ty)
	if !ok {
		tyName = fmt.Sprintf("type %d", ty)
	}
	fmt.Fprintf(&b, "id: %s\n", id)
	fmt.Fprintf(&b, "type: %s\n", tyName)

	switch stored := stored.(type) {
	case *objectstored.StoredBlob:
		blob := stored.Blob()
		fmt.Fprintf(&b, "size: %d\n", len(blob.Data))
		fmt.Fprintf(&b, "data: %q\n", string(blob.Data))
	case *objectstored.StoredTree:
		tree := stored.Tree()
		fmt.Fprintf(&b, "entries: %d\n", len(tree.Entries))
		for _, entry := range tree.Entries {
			fmt.Fprintf(&b, "%06o %s\t%s\n", entry.Mode, entry.ID, entry.Name)
		}
	case *objectstored.StoredCommit:
		commit := stored.Commit()
		fmt.Fprintf(&b, "tree: %s\n", commit.Tree)
		for _, parent := range commit.Parents {
			fmt.Fprintf(&b, "parent: %s\n", parent)
		}
		fmt.Fprintf(&b, "author: %s <%s>\n", commit.Author.Name, commit.Author.Email)
		fmt.Fprintf(&b, "committer: %s <%s>\n", commit.Committer.Name, commit.Committer.Email)
		fmt.Fprintf(&b, "message:\n%s\n", string(commit.Message))
	case *objectstored.StoredTag:
		tag := stored.Tag()
		targetTy, ok := objecttype.Name(tag.TargetType)
		if !ok {
			targetTy = fmt.Sprintf("type %d", tag.TargetType)
		}
		fmt.Fprintf(&b, "target: %s (%s)\n", tag.Target, targetTy)
		fmt.Fprintf(&b, "name: %s\n", tag.Name)
		if tag.Tagger != nil {
			fmt.Fprintf(&b, "tagger: %s <%s>\n", tag.Tagger.Name, tag.Tagger.Email)
		}
		fmt.Fprintf(&b, "message:\n%s\n", string(tag.Message))
	default:
		fmt.Fprintf(&b, "%#v\n", stored.Object())
	}

	_, _ = os.Stdout.WriteString(b.String())
}
