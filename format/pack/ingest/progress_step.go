package ingest

func progressStep(total uint32) uint32 {
	if total <= 200 {
		return 1
	}

	return total / 200
}
