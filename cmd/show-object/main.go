// Command show-object provides a small command line utility to show the details of a specified Git object.
package main

import (
	"flag"
	"log"
)

func main() {
	repoPath := flag.String("r", "", "path to git dir (.git or bare repo root)")
	name := flag.String("h", "", "reference name or object id")

	flag.Parse()

	if *repoPath == "" || *name == "" {
		log.Fatal("must provide -r <repo> and -h <ref-or-object-id>")
	}

	err := run(repoPath, name)
	if err != nil {
		log.Fatalf("run: %v", err)
	}
}
