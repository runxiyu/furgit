package service

import (
	"context"
	"os"

	"codeberg.org/lindenii/furgit/internal/utils"
	objectstore "codeberg.org/lindenii/furgit/object/store"
	"codeberg.org/lindenii/furgit/object/store/loose"
	objectmix "codeberg.org/lindenii/furgit/object/store/mix"
	"codeberg.org/lindenii/furgit/object/store/packed"
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

	utils.BestEffortFprintf(service.opts.Progress, "running hooks...\r")

	quarantinedObjects := service.opts.ExistingObjects

	var (
		quarantineObjectsStore objectstore.ReadingStore
		quarantineLooseStore   *loose.Store
		quarantinePackedStore  *packed.Store
		quarantineLooseRoot    *os.Root
		quarantinePackRoot     *os.Root
		err                    error
	)

	//nolint:nestif
	if quarantineName != "" {
		quarantineLooseRoot, err = service.opts.ObjectsRoot.OpenRoot(quarantineName)
		if err != nil {
			utils.BestEffortFprintf(service.opts.Progress, "running hooks: failed: %v.\n", err)

			return nil, nil, nil, false, err.Error()
		}

		quarantineLooseStore, err = loose.New(quarantineLooseRoot, service.opts.Algorithm)
		if err != nil {
			_ = quarantineLooseRoot.Close()

			utils.BestEffortFprintf(service.opts.Progress, "running hooks: failed: %v.\n", err)

			return nil, nil, nil, false, err.Error()
		}

		quarantineObjectsStore = quarantineLooseStore
		quarantinedObjects = quarantineLooseStore

		quarantinePackRoot, err = quarantineLooseRoot.OpenRoot("pack")
		if err == nil {
			var packedErr error

			quarantinePackedStore, packedErr = packed.New(quarantinePackRoot, service.opts.Algorithm, packed.Options{})
			if packedErr != nil {
				_ = quarantineLooseStore.Close()
				_ = quarantinePackRoot.Close()
				_ = quarantineLooseRoot.Close()

				utils.BestEffortFprintf(service.opts.Progress, "running hooks: failed: %v.\n", packedErr)

				return nil, nil, nil, false, packedErr.Error()
			}

			quarantineObjectsStore = objectmix.New(quarantineLooseStore, quarantinePackedStore)
			quarantinedObjects = quarantineObjectsStore
		} else if !os.IsNotExist(err) {
			_ = quarantineLooseStore.Close()
			_ = quarantineLooseRoot.Close()

			utils.BestEffortFprintf(service.opts.Progress, "running hooks: failed: %v.\n", err)

			return nil, nil, nil, false, err.Error()
		}

		defer func() {
			if quarantineObjectsStore != nil {
				_ = quarantineObjectsStore.Close()
			}

			if quarantinePackedStore != nil {
				_ = quarantinePackedStore.Close()
			}

			if quarantineLooseStore != nil {
				_ = quarantineLooseStore.Close()
			}

			if quarantinePackRoot != nil {
				_ = quarantinePackRoot.Close()
			}

			if quarantineLooseRoot != nil {
				_ = quarantineLooseRoot.Close()
			}
		}()
	}

	decisions, err := service.opts.Hook(ctx, HookRequest{
		Refs:               service.opts.Refs,
		ExistingObjects:    service.opts.ExistingObjects,
		QuarantinedObjects: quarantinedObjects,
		Updates:            buildHookUpdates(commands),
		PushOptions:        append([]string(nil), req.PushOptions...),
		IO:                 service.opts.HookIO,
	})
	if err != nil {
		utils.BestEffortFprintf(service.opts.Progress, "running hooks: failed: %v.\n", err)

		return nil, nil, nil, false, err.Error()
	}

	if len(decisions) != len(commands) {
		utils.BestEffortFprintf(service.opts.Progress, "running hooks: failed: wrong decision count.\n")

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

	utils.BestEffortFprintf(
		service.opts.Progress,
		"running hooks: done (%d/%d accepted).\n",
		len(allowedCommands),
		len(commands),
	)

	return allowedCommands, allowedIndices, rejected, true, ""
}
