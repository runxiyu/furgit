// Command receivepack9418 serves one fixed repository over git:// receive-pack on TCP 9418.
package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	os.Exit(runMain())
}

func runMain() int {
	listenAddr := flag.String("listen", ":9418", "listen address")
	repoPath := flag.String("repo", "", "path to git dir (.git or bare repo root)")
	cpuProfilePath := flag.String("cpuprofile", "", "write CPU profile to file")
	memProfilePath := flag.String("memprofile", "", "write heap profile to file at exit")

	flag.Parse()

	if *repoPath == "" {
		log.Print("must provide -repo <path-to-git-dir>")

		return 2
	}

	if *cpuProfilePath != "" {
		stopCPUProfile, err := startCPUProfile(*cpuProfilePath)
		if err != nil {
			log.Printf("cpuprofile: %v", err)

			return 1
		}

		defer func() {
			stopErr := stopCPUProfile()
			if stopErr != nil {
				log.Printf("cpuprofile: %v", stopErr)
			}
		}()
	}

	if *memProfilePath != "" {
		defer func() {
			memErr := writeMemProfile(*memProfilePath)
			if memErr != nil {
				log.Printf("memprofile: %v", memErr)
			}
		}()
	}

	err := run(*listenAddr, *repoPath)
	if err != nil {
		log.Printf("run: %v", err)

		return 1
	}

	return 0
}
