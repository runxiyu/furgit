package refname

func isCurrentWorktreeRef(name string) bool {
	return IsRootSyntax(name) || IsPerWorktree(name)
}
