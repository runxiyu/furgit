package service

import objectid "codeberg.org/lindenii/furgit/object/id"

// Command is one protocol-independent requested ref update.
type Command struct {
	OldID objectid.ObjectID
	NewID objectid.ObjectID
	Name  string
}

func fillCommandErrors(result *Result, commands []Command, errText string) {
	for _, command := range commands {
		result.Commands = append(result.Commands, CommandResult{
			Name:    command.Name,
			Error:   errText,
			RefName: command.Name,
			OldID:   new(command.OldID),
			NewID:   new(command.NewID),
		})
	}
}

func isDelete(command Command) bool {
	return command.NewID == command.NewID.Algorithm().Zero()
}
