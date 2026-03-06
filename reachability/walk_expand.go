package reachability

func (walk *Walk) expand(item walkItem) ([]walkItem, error) {
	if walk.domain == DomainCommits {
		return walk.expandCommits(item)
	}

	return walk.expandObjects(item)
}
