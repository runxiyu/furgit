package service

import (
	"context"
	"log"

	"codeberg.org/lindenii/furgit/format/pack/ingest"
)

// Execute validates one receive-pack request, optionally ingests its pack into
// quarantine, and plans ref updates.
//
// TODO: Invoke hook or policy callbacks to decide whether each planned update
// should be allowed.
// TODO: Apply planned ref updates with one atomic compare-and-swap ref
// transaction once ref writing exists.
func (service *Service) Execute(ctx context.Context, req *Request) (*Result, error) {
	_ = ctx

	result := &Result{
		Commands: make([]CommandResult, 0, len(req.Commands)),
	}

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

		quarantineName, quarantineRoot, err := service.createQuarantineRoot()
		if err != nil {
			result.UnpackError = err.Error()
			fillCommandErrors(result, req.Commands, err.Error())

			return result, nil
		}

		defer func() {
			_ = quarantineRoot.Close()
			// TODO: Promote accepted quarantined objects into the permanent object
			// store once atomic ref application exists.
			_ = service.opts.ObjectsRoot.RemoveAll(quarantineName)
		}()

		ingested, err := ingest.Ingest(
			req.Pack,
			quarantineRoot,
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

	fillCommandErrors(result, req.Commands, "ref updates not implemented yet")
	log.Printf(
		"receivepack: planned %d ref updates, but hook/policy checks and atomic ref writes are not implemented yet",
		len(result.Planned),
	)

	return result, nil
}
