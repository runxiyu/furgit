package name

import "strings"

// ValidateUpdateName checks whether name is valid for a direct ref update.
func ValidateUpdateName(name string, hasNewValue bool) error {
	if IsPseudo(name) {
		return &NameError{Name: name, Reason: "pseudoref updates are not allowed"}
	}

	if hasNewValue {
		return Validate(name, Options{AllowOneLevel: true, RefspecPattern: false})
	}

	if !IsSafe(name) {
		return &NameError{Name: name, Reason: "unsafe name for update"}
	}

	return nil
}

// ValidateSymbolicTarget checks whether target is valid for a symref target.
func ValidateSymbolicTarget(name string, target string) error {
	parsed := ParseWorktree(name)
	if parsed.BareRefName == "HEAD" && !strings.HasPrefix(target, "refs/heads/") {
		return &NameError{Name: target, Reason: name + " must point to refs/heads/..."}
	}

	if IsRoot(target) {
		return nil
	}

	err := Validate(target, Options{AllowOneLevel: false, RefspecPattern: false})
	if err != nil {
		return err
	}

	if strings.HasPrefix(target, "refs/") {
		return nil
	}

	if strings.HasPrefix(target, "worktrees/") {
		return nil
	}

	return &NameError{Name: target, Reason: "symref target is not a ref"}
}
