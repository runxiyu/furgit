package refname

import "strings"

// ValidateUpdateName checks whether name is valid for one direct ref update.
//
// See transaction_refname_valid();
// updates with a new OID use check_refname_format(..., ALLOW_ONELEVEL),
// while delete/verify style operations use refname_is_safe().
func ValidateUpdateName(name string, hasNewValue bool) error {
	if IsPseudo(name) {
		return &NameError{Name: name, Reason: "pseudoref updates are not allowed"}
	}

	if hasNewValue {
		return Validate(name, Options{AllowOneLevel: true})
	}

	if !IsSafe(name) {
		return &NameError{Name: name, Reason: "unsafe refname for update"}
	}

	return nil
}

// ValidateSymbolicTarget checks whether target is valid for one symref target.
//
// See refs_fsck_symref();
// root refs are allowed directly, HEAD must point to refs/heads/...,
// and non-root targets must be valid full refnames rooted at refs/ or
// worktrees/.
func ValidateSymbolicTarget(refname string, target string) error {
	parsed := ParseWorktree(refname)
	if parsed.BareRefName == "HEAD" && !strings.HasPrefix(target, "refs/heads/") {
		return &NameError{Name: target, Reason: refname + " must point to refs/heads/..."}
	}

	if IsRoot(target) {
		return nil
	}

	err := Validate(target, Options{})
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
