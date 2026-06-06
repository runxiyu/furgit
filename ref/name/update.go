package name

import (
	"fmt"
	"strings"
)

// ValidateUpdateName checks whether name is valid for a direct ref update.
func ValidateUpdateName(name string, hasNewValue bool) error {
	if IsPseudo(name) {
		return fmt.Errorf("%w: pseudoref updates are not allowed", ErrInvalidName)
	}

	if hasNewValue {
		return Validate(name, Options{AllowOneLevel: true, RefspecPattern: false})
	}

	if !IsSafe(name) {
		return fmt.Errorf("%w: unsafe name for update", ErrInvalidName)
	}

	return nil
}

// ValidateSymbolicTarget checks whether target is valid for a symref target.
func ValidateSymbolicTarget(name string, target string) error {
	parsed := ParseWorktree(name)
	if parsed.BareRefName == "HEAD" && !strings.HasPrefix(target, "refs/heads/") {
		return fmt.Errorf("%w: %s must point to refs/heads/...", ErrInvalidName, name)
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

	return fmt.Errorf("%w: symref target is not a ref", ErrInvalidName)
}
