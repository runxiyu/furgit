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
			OldID:   objectIDPointer(command.OldID),
			NewID:   objectIDPointer(command.NewID),
		})
	}
}

func isDelete(command Command) bool {
	return command.NewID == objectid.Zero(command.NewID.Algorithm())
}

func objectIDPointer(id objectid.ObjectID) *objectid.ObjectID {
	out := id

	return &out
}
