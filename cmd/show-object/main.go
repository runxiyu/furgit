package main

import (
	"flag"
	"fmt"
	"log"

	"codeberg.org/lindenii/furgit"
)

func main() {
	repoPath := flag.String("r", "", "path to repo (.git or bare)")
	ref := flag.String("h", "", "ref or hash")
	flag.Parse()

	if *repoPath == "" || *ref == "" {
		log.Fatal("must provide -r repo and -h ref/hash")
	}

	repo, err := furgit.OpenRepository(*repoPath)
	if err != nil {
		log.Fatalf("open repo: %v", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	h, err := repo.ResolveRefFully(*ref)
	if err != nil {
		log.Fatalf("resolve ref: %v", err)
	}

	obj, err := repo.ReadObject(h.Hash)
	if err != nil {
		log.Fatalf("read object: %v", err)
	}

	fmt.Printf("%#v\n", obj)
}
