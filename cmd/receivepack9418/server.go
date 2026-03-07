package main

import (
	"os"

	"codeberg.org/lindenii/furgit/repository"
)

type server struct {
	repo        *repository.Repository
	objectsRoot *os.Root
}
