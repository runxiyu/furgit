package service

import (
	"codeberg.org/lindenii/furgit/internal/utils"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/refstore"
)

func (service *Service) applyAtomic(result *Result, commands []Command) error {
	total := len(commands)
	utils.WriteProgressf(service.opts.Progress, "updating refs: 0/%d\r", total)

	tx, err := service.opts.Refs.BeginTransaction()
	if err != nil {
		return err
	}

	for i, command := range commands {
		err = queueWriteTransaction(tx, command)
		if err != nil {
			_ = tx.Abort()

			fillCommandErrors(result, commands, err.Error())
			utils.WriteProgressf(service.opts.Progress, "updating refs: failed at %d/%d\n", i+1, total)

			return nil
		}

		utils.WriteProgressf(service.opts.Progress, "updating refs: %d/%d\r", i+1, total)
	}

	err = tx.Commit()
	if err != nil {
		fillCommandErrors(result, commands, err.Error())
		utils.WriteProgressf(service.opts.Progress, "updating refs: failed at commit\n")

		return nil
	}

	result.Applied = true
	for _, command := range commands {
		result.Commands = append(result.Commands, successCommandResult(command))
	}
	utils.WriteProgressf(service.opts.Progress, "updating refs: done.\n")

	return nil
}

func (service *Service) applyBatch(result *Result, commands []Command) error {
	total := len(commands)
	utils.WriteProgressf(service.opts.Progress, "updating refs...\r")

	batch, err := service.opts.Refs.BeginBatch()
	if err != nil {
		return err
	}

	for _, command := range commands {
		queueWriteBatch(batch, command)
	}

	batchResults, err := batch.Apply()
	if err != nil && len(batchResults) == 0 {
		utils.WriteProgressf(service.opts.Progress, "updating refs: failed at apply\n")

		return err
	}

	appliedAny := false

	for i, command := range commands {
		item := successCommandResult(command)
		if i < len(batchResults) && batchResults[i].Error != nil {
			item.Error = batchResults[i].Error.Error()
		} else {
			appliedAny = true
		}

		result.Commands = append(result.Commands, item)
		utils.WriteProgressf(service.opts.Progress, "updating refs: %d/%d\r", i+1, total)
	}

	result.Applied = appliedAny
	utils.WriteProgressf(service.opts.Progress, "updating refs: done.\n")

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
