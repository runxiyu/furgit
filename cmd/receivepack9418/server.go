package main

import (
	"os"

	"lindenii.org/go/furgit/repository"
)

type server struct {
	repo        *repository.Repository
	objectsRoot *os.Root
}
