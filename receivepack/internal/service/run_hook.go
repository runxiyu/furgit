package service

import "context"

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

	quarantinedObjects, err := service.openQuarantinedObjects(quarantineName)
	if err != nil {
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
	})
	if err != nil {
		return nil, nil, nil, false, err.Error()
	}

	if len(decisions) != len(commands) {
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

	return allowedCommands, allowedIndices, rejected, true, ""
}
