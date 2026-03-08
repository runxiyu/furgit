package service

import (
	"context"

	"codeberg.org/lindenii/furgit/internal/utils"
)

func (service *Service) runHook(
	ctx context.Context,
	req *Request,
	commands []Command,
	quarantineName string,
) (
	allowedCommands []Command,
	allowedIndices []int,
	rejected map[int]string,
	ok bool,
	errText string,
) {
	allowedCommands = append([]Command(nil), commands...)

	allowedIndices = make([]int, 0, len(commands))
	for index := range commands {
		allowedIndices = append(allowedIndices, index)
	}

	rejected = make(map[int]string)
	if service.opts.Hook == nil {
		return allowedCommands, allowedIndices, rejected, true, ""
	}

	utils.FprintfBestEffort(service.opts.Progress, "running hooks...\r")

	quarantinedObjects, err := service.openQuarantinedObjects(quarantineName)
	if err != nil {
		utils.FprintfBestEffort(service.opts.Progress, "running hooks: failed: %v.\n", err)

		return nil, nil, nil, false, err.Error()
	}

	defer func() {
		_ = quarantinedObjects.Close()
	}()

	decisions, err := service.opts.Hook(ctx, HookRequest{
		Refs:               service.opts.Refs,
		ExistingObjects:    service.opts.ExistingObjects,
		QuarantinedObjects: quarantinedObjects,
		Updates:            buildHookUpdates(commands),
		PushOptions:        append([]string(nil), req.PushOptions...),
		IO:                 service.opts.HookIO,
	})
	if err != nil {
		utils.FprintfBestEffort(service.opts.Progress, "running hooks: failed: %v.\n", err)

		return nil, nil, nil, false, err.Error()
	}

	if len(decisions) != len(commands) {
		utils.FprintfBestEffort(service.opts.Progress, "running hooks: failed: wrong decision count.\n")

		return nil, nil, nil, false, "hook returned wrong number of update decisions"
	}

	allowedCommands = allowedCommands[:0]
	allowedIndices = allowedIndices[:0]

	for index, decision := range decisions {
		if decision.Accept {
			allowedCommands = append(allowedCommands, commands[index])
			allowedIndices = append(allowedIndices, index)

			continue
		}

		message := decision.Message
		if message == "" {
			message = "rejected by hook"
		}

		rejected[index] = message
	}

	utils.FprintfBestEffort(
		service.opts.Progress,
		"running hooks: done (%d/%d accepted).\n",
		len(allowedCommands),
		len(commands),
	)

	return allowedCommands, allowedIndices, rejected, true, ""
}
