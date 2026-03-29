// Command index-pack ingests one pack stream from stdin and writes .pack/.idx/.rev.
package main

import (
	"flag"
	"log"
)

func main() {
	repoPath := flag.String("r", "", "path to git dir (.git or bare repo root)")
	destinationPath := flag.String("destination", "", "path to destination objects/pack directory")
	objectFormat := flag.String("object-format", "", "object format (sha1 or sha256)")
	fixThin := flag.Bool("fix-thin", false, "fix thin packs using repository object store")
	writeRev := flag.Bool("rev-index", true, "write reverse index (.rev)")

	flag.Parse()

	if *destinationPath == "" {
		log.Fatal("must provide -destination <objects/pack>")
	}

	err := run(*repoPath, *destinationPath, *objectFormat, *fixThin, *writeRev)
	if err != nil {
		log.Fatalf("run: %v", err)
	}
}
