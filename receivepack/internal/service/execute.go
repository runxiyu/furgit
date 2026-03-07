package service

import (
	"context"
	"os"

	"codeberg.org/lindenii/furgit/format/pack/ingest"
)

// Execute validates one receive-pack request, optionally ingests its pack into
// quarantine, and plans ref updates.
//
// TODO: Invoke hook or policy callbacks to decide whether each planned update
// should be allowed.
func (service *Service) Execute(ctx context.Context, req *Request) (*Result, error) {
	_ = ctx

	result := &Result{
		Commands: make([]CommandResult, 0, len(req.Commands)),
	}
	var (
		quarantineName string
		quarantineRoot *os.Root
		err            error
	)

	if req.PackExpected {
		if req.Pack == nil {
			result.UnpackError = "missing pack stream"
			fillCommandErrors(result, req.Commands, "missing pack stream")

			return result, nil
		}

		if service.opts.ObjectsRoot == nil {
			result.UnpackError = "objects root not configured"
			fillCommandErrors(result, req.Commands, "objects root not configured")

			return result, nil
		}

		quarantineName, quarantineRoot, err = service.createQuarantineRoot()
		if err != nil {
			result.UnpackError = err.Error()
			fillCommandErrors(result, req.Commands, err.Error())

			return result, nil
		}

		defer func() {
			_ = quarantineRoot.Close()
			_ = service.opts.ObjectsRoot.RemoveAll(quarantineName)
		}()

		quarantinePackRoot, err := service.openQuarantinePackRoot(quarantineRoot)
		if err != nil {
			result.UnpackError = err.Error()
			fillCommandErrors(result, req.Commands, err.Error())

			return result, nil
		}

		defer func() {
			_ = quarantinePackRoot.Close()
		}()

		ingested, err := ingest.Ingest(
			req.Pack,
			quarantinePackRoot,
			service.opts.Algorithm,
			true,
			true,
			service.opts.ExistingObjects,
		)
		if err != nil {
			result.UnpackError = err.Error()
			fillCommandErrors(result, req.Commands, err.Error())

			return result, nil
		}

		result.Ingest = &ingested
	}

	for _, command := range req.Commands {
		result.Planned = append(result.Planned, PlannedUpdate{
			Name:   command.Name,
			OldID:  command.OldID,
			NewID:  command.NewID,
			Delete: isDelete(command),
		})
	}

	if len(req.Commands) == 0 {
		return result, nil
	}

	if req.PackExpected {
		// Git migrates quarantined objects into permanent storage immediately
		// before starting ref updates.
		err = service.promoteQuarantine(quarantineName, quarantineRoot)
		if err != nil {
			result.UnpackError = err.Error()
			fillCommandErrors(result, req.Commands, err.Error())

			return result, nil
		}
	}

	if req.Atomic {
		err := service.applyAtomic(result, req.Commands)
		if err != nil {
			return result, err
		}

		return result, nil
	}

	err = service.applyBatch(result, req.Commands)
	if err != nil {
		return result, err
	}

	return result, nil
}
