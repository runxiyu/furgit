package service

import (
	"context"

	"codeberg.org/lindenii/furgit/internal/utils"
	objectstore "codeberg.org/lindenii/furgit/object/store"
)

// Execute validates one receive-pack request, optionally ingests its pack into
// quarantine, runs the optional hook, and applies allowed ref updates.
//
// Labels: Deps-Borrowed.
func (service *Service) Execute(ctx context.Context, req *Request) (*Result, error) {
	result := &Result{
		Commands: make([]CommandResult, 0, len(req.Commands)),
	}
	var err error

	quarantine, ok := service.ingestQuarantine(result, req.Commands, req)
	if !ok {
		return result, nil
	}

	if quarantine != nil {
		defer func(q objectstore.Quarantine) {
			_ = q.Discard()
		}(quarantine)
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
		quarantine,
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

	if req.PackExpected && quarantine != nil {
		// Git migrates quarantined objects into permanent storage immediately
		// before starting ref updates.
		utils.BestEffortFprintf(service.opts.Progress, "promoting quarantine...\r")

		err := quarantine.Promote()
		if err != nil {
			utils.BestEffortFprintf(service.opts.Progress, "promoting quarantine: failed: %v.\n", err)

			result.UnpackError = err.Error()
			fillCommandErrors(result, req.Commands, err.Error())

			return result, nil
		}

		quarantine = nil
		utils.BestEffortFprintf(service.opts.Progress, "promoting quarantine: done.\n")
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
