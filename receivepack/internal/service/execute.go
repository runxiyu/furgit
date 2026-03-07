package service

import (
	"context"
	"os"

	"codeberg.org/lindenii/furgit/format/pack/ingest"
)

// Execute validates one receive-pack request, optionally ingests its pack into
// quarantine, runs the optional hook, and applies allowed ref updates.
func (service *Service) Execute(ctx context.Context, req *Request) (*Result, error) {
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

	allowedCommands := append([]Command(nil), req.Commands...)
	allowedIndices := make([]int, 0, len(req.Commands))
	for index := range req.Commands {
		allowedIndices = append(allowedIndices, index)
	}
	rejected := make(map[int]string)

	if service.opts.Hook != nil {
		quarantinedObjects, err := service.openQuarantinedObjects(quarantineName)
		if err != nil {
			fillCommandErrors(result, req.Commands, err.Error())

			return result, nil
		}

		defer func() {
			_ = quarantinedObjects.Close()
		}()

		decisions, err := service.opts.Hook(ctx, HookRequest{
			Refs:               service.opts.Refs,
			ExistingObjects:    service.opts.ExistingObjects,
			QuarantinedObjects: quarantinedObjects,
			Updates:            buildHookUpdates(req.Commands),
			PushOptions:        append([]string(nil), req.PushOptions...),
		})
		if err != nil {
			fillCommandErrors(result, req.Commands, err.Error())

			return result, nil
		}

		if len(decisions) != len(req.Commands) {
			fillCommandErrors(result, req.Commands, "hook returned wrong number of update decisions")

			return result, nil
		}

		allowedCommands = allowedCommands[:0]
		allowedIndices = allowedIndices[:0]
		for index, decision := range decisions {
			if decision.Accept {
				allowedCommands = append(allowedCommands, req.Commands[index])
				allowedIndices = append(allowedIndices, index)

				continue
			}

			message := decision.Message
			if message == "" {
				message = "rejected by hook"
			}

			rejected[index] = message
		}

		if req.Atomic && len(rejected) != 0 {
			result.Commands = make([]CommandResult, 0, len(req.Commands))
			for index, command := range req.Commands {
				message := rejected[index]
				if message == "" {
					message = "atomic push rejected by hook"
				}

				result.Commands = append(result.Commands, resultForHookRejection(command, message))
			}

			return result, nil
		}
	}

	if len(allowedCommands) == 0 {
		result.Commands = mergeCommandResults(req.Commands, rejected, nil, nil)

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
		subresult := &Result{}
		err := service.applyAtomic(subresult, allowedCommands)
		if err != nil {
			return result, err
		}

		result.Commands = mergeCommandResults(req.Commands, rejected, subresult.Commands, allowedIndices)
		result.Applied = subresult.Applied

		return result, nil
	}

	subresult := &Result{}
	err = service.applyBatch(subresult, allowedCommands)
	if err != nil {
		return result, err
	}

	result.Commands = mergeCommandResults(req.Commands, rejected, subresult.Commands, allowedIndices)
	result.Applied = subresult.Applied

	return result, nil
}
