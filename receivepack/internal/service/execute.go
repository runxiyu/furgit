package service

import (
	"context"
	"os"
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

	quarantineName, quarantineRoot, ok := service.ingestQuarantine(result, req.Commands, req)
	if !ok {
		return result, nil
	}

	if quarantineRoot != nil {
		defer func() {
			_ = quarantineRoot.Close()
			_ = service.opts.ObjectsRoot.RemoveAll(quarantineName)
		}()
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

	allowedCommands, allowedIndices, rejected, ok, errText := service.runHook(
		ctx,
		req,
		req.Commands,
		quarantineName,
	)
	if !ok {
		fillCommandErrors(result, req.Commands, errText)

		return result, nil
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
