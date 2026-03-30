package service

func resultForHookRejection(command Command, message string) CommandResult {
	result := successCommandResult(command)
	result.Error = message

	return result
}

func mergeCommandResults(
	commands []Command,
	rejected map[int]string,
	applied []CommandResult,
	appliedIndices []int,
) []CommandResult {
	out := make([]CommandResult, len(commands))

	for index, message := range rejected {
		out[index] = resultForHookRejection(commands[index], message)
	}

	for i, appliedResult := range applied {
		if i >= len(appliedIndices) {
			break
		}

		out[appliedIndices[i]] = appliedResult
	}

	return out
}
