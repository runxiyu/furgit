package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	"codeberg.org/lindenii/furgit/repository"
)

func run(listenAddr, repoPath string) error {
	repoRoot, err := os.OpenRoot(repoPath)
	if err != nil {
		return fmt.Errorf("open repo root: %w", err)
	}

	defer func() { _ = repoRoot.Close() }()

	repo, err := repository.Open(repoRoot)
	if err != nil {
		return fmt.Errorf("open repository: %w", err)
	}

	defer func() { _ = repo.Close() }()

	objectsRoot, err := repoRoot.OpenRoot("objects")
	if err != nil {
		return fmt.Errorf("open objects root: %w", err)
	}

	defer func() { _ = objectsRoot.Close() }()

	srv := &server{
		repo:        repo,
		objectsRoot: objectsRoot,
	}

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("listen %q: %w", listenAddr, err)
	}

	defer func() { _ = ln.Close() }()

	log.Printf("receivepack9418: listening on %s", listenAddr)
	log.Printf("receivepack9418: repository=%s algorithm=%s", repoPath, repo.Algorithm())

	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}

			var nerr net.Error
			if errors.As(err, &nerr) && nerr.Temporary() {
				log.Printf("receivepack9418: temporary accept error: %v", err)

				continue
			}

			return fmt.Errorf("accept: %w", err)
		}

		go srv.handleConn(conn)
	}
}
