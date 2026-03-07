package service

import (
	"context"
	"io"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	"codeberg.org/lindenii/furgit/refstore"
)

type HookIO struct {
	Progress io.Writer
	Error    io.Writer
}

type RefUpdate struct {
	Name  string
	OldID objectid.ObjectID
	NewID objectid.ObjectID
}

type UpdateDecision struct {
	Accept  bool
	Message string
}

type HookRequest struct {
	Refs               refstore.ReadingStore
	ExistingObjects    objectstore.Store
	QuarantinedObjects objectstore.Store
	Updates            []RefUpdate
	PushOptions        []string
	IO                 HookIO
}

type Hook func(context.Context, HookRequest) ([]UpdateDecision, error)
