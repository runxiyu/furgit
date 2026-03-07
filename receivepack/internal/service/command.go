package service

import "codeberg.org/lindenii/furgit/objectid"

// Command is one protocol-independent requested ref update.
type Command struct {
	OldID objectid.ObjectID
	NewID objectid.ObjectID
	Name  string
}

func fillCommandErrors(result *Result, commands []Command, errText string) {
	for _, command := range commands {
		result.Commands = append(result.Commands, CommandResult{
			Name:  command.Name,
			Error: errText,
		})
	}
}

func isDelete(command Command) bool {
	return command.NewID == objectid.Zero(command.NewID.Algorithm())
}
