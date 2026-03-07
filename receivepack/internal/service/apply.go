package service

import (
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/refstore"
)

func (service *Service) applyAtomic(result *Result, commands []Command) error {
	tx, err := service.opts.Refs.BeginTransaction()
	if err != nil {
		return err
	}

	for _, command := range commands {
		err = queueWriteTransaction(tx, command)
		if err != nil {
			_ = tx.Abort()

			fillCommandErrors(result, commands, err.Error())

			return nil
		}
	}

	err = tx.Commit()
	if err != nil {
		fillCommandErrors(result, commands, err.Error())

		return nil
	}

	result.Applied = true
	for _, command := range commands {
		result.Commands = append(result.Commands, successCommandResult(command))
	}

	return nil
}

func (service *Service) applyBatch(result *Result, commands []Command) error {
	batch, err := service.opts.Refs.BeginBatch()
	if err != nil {
		return err
	}

	for _, command := range commands {
		queueWriteBatch(batch, command)
	}

	batchResults, err := batch.Apply()
	if err != nil && len(batchResults) == 0 {
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
	}

	result.Applied = appliedAny

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
