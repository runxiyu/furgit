// Command receivepack9418 serves one fixed repository over git:// receive-pack on TCP 9418.
package main

import (
	"flag"
	"log"
)

func main() {
	listenAddr := flag.String("listen", ":9418", "listen address")
	repoPath := flag.String("repo", "", "path to git dir (.git or bare repo root)")

	flag.Parse()

	if *repoPath == "" {
		log.Fatal("must provide -repo <path-to-git-dir>")
	}

	err := run(*listenAddr, *repoPath)
	if err != nil {
		log.Fatalf("run: %v", err)
	}
}
