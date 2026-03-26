package service

import (
	"codeberg.org/lindenii/furgit/internal/utils"
	objectid "codeberg.org/lindenii/furgit/object/id"
	refstore "codeberg.org/lindenii/furgit/ref/store"
)

func (service *Service) applyAtomic(result *Result, commands []Command) error {
	total := len(commands)
	utils.BestEffortFprintf(service.opts.Progress, "updating refs: 0/%d\r", total)

	tx, err := service.opts.Refs.BeginTransaction()
	if err != nil {
		return err
	}

	for i, command := range commands {
		err = queueWriteTransaction(tx, command)
		if err != nil {
			_ = tx.Abort()

			fillCommandErrors(result, commands, err.Error())
			utils.BestEffortFprintf(service.opts.Progress, "updating refs: failed at %d/%d.\n", i+1, total)

			return nil
		}

		utils.BestEffortFprintf(service.opts.Progress, "updating refs: %d/%d\r", i+1, total)
	}

	err = tx.Commit()
	if err != nil {
		fillCommandErrors(result, commands, err.Error())
		utils.BestEffortFprintf(service.opts.Progress, "updating refs: failed at commit.\n")

		return nil
	}

	result.Applied = true
	for _, command := range commands {
		result.Commands = append(result.Commands, successCommandResult(command))
	}

	utils.BestEffortFprintf(service.opts.Progress, "updating refs: done.\n")

	return nil
}

func (service *Service) applyBatch(result *Result, commands []Command) error {
	total := len(commands)

	utils.BestEffortFprintf(service.opts.Progress, "updating refs...\r")

	batch, err := service.opts.Refs.BeginBatch()
	if err != nil {
		return err
	}

	for _, command := range commands {
		queueWriteBatch(batch, command)
	}

	batchResults, err := batch.Apply()
	if err != nil && len(batchResults) == 0 {
		utils.BestEffortFprintf(service.opts.Progress, "updating refs: failed at apply.\n")

		return err
	}

	appliedAny := false
	failedCount := 0

	for i, command := range commands {
		item := successCommandResult(command)
		if i < len(batchResults) && batchResults[i].Error != nil {
			item.Error = batchResults[i].Error.Error()
			failedCount++
		} else {
			appliedAny = true
		}

		result.Commands = append(result.Commands, item)

		utils.BestEffortFprintf(service.opts.Progress, "updating refs: %d/%d\r", i+1, total)
	}

	result.Applied = appliedAny

	if failedCount == 0 {
		utils.BestEffortFprintf(service.opts.Progress, "updating refs: done.\n")
	} else {
		utils.BestEffortFprintf(service.opts.Progress, "updating refs: failed (%d/%d).\n", failedCount, total)
	}

	return nil
}

func queueWriteTransaction(tx refstore.Transaction, command Command) error {
	if isDelete(command) {
		return tx.Delete(command.Name, command.OldID)
	}

	if command.OldID == objectid.Zero(command.OldID.Algorithm()) {
		return tx.Create(command.Name, command.NewID)
	}

	return tx.Update(command.Name, command.NewID, command.OldID)
}

func queueWriteBatch(batch refstore.Batch, command Command) {
	if isDelete(command) {
		batch.Delete(command.Name, command.OldID)

		return
	}

	if command.OldID == objectid.Zero(command.OldID.Algorithm()) {
		batch.Create(command.Name, command.NewID)

		return
	}

	batch.Update(command.Name, command.NewID, command.OldID)
}

func successCommandResult(command Command) CommandResult {
	return CommandResult{
		Name:    command.Name,
		RefName: command.Name,
		OldID:   objectIDPointer(command.OldID),
		NewID:   objectIDPointer(command.NewID),
	}
}
