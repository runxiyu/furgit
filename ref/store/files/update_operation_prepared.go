package files

type preparedUpdate struct {
	op     queuedUpdate
	target resolvedUpdateTarget
}
