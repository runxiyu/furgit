package name_test

import (
	"testing"

	"lindenii.org/go/furgit/ref/name"
)

func TestValidateUpdateName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		hasNewValue bool
		wantErr     bool
	}{
		{name: "refs/heads/main", hasNewValue: true, wantErr: false},
		{name: "HEAD", hasNewValue: true, wantErr: false},
		{name: "PSEUDOREF", hasNewValue: true, wantErr: false},
		{name: "FETCH_HEAD", hasNewValue: true, wantErr: true},
		{name: "MERGE_HEAD", hasNewValue: true, wantErr: true},
		{name: "refs/heads/.bad", hasNewValue: true, wantErr: true},
		{name: "foo/bar", hasNewValue: true, wantErr: false},
		{name: "foo/bar", hasNewValue: false, wantErr: true},
		{name: "PSEUDOREF", hasNewValue: false, wantErr: false},
		{name: "HEAD", hasNewValue: false, wantErr: false},
	}

	for _, tt := range tests {
		err := name.ValidateUpdateName(tt.name, tt.hasNewValue)
		if (err != nil) != tt.wantErr {
			t.Fatalf("ValidateUpdateName(%q, %v) err=%v, wantErr=%v", tt.name, tt.hasNewValue, err, tt.wantErr)
		}
	}
}

func TestValidateSymbolicTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		ref     string
		target  string
		wantErr bool
	}{
		{ref: "HEAD", target: "refs/heads/main", wantErr: false},
		{ref: "HEAD", target: "foo", wantErr: true},
		{ref: "HEAD", target: "ORIG_HEAD", wantErr: true},
		{ref: "refs/heads/top", target: "ORIG_HEAD", wantErr: false},
		{ref: "refs/heads/top", target: "refs/heads/main", wantErr: false},
		{ref: "refs/heads/top", target: "worktrees/wt1/HEAD", wantErr: false},
		{ref: "refs/heads/top", target: "foo", wantErr: true},
		{ref: "refs/heads/top", target: "foo..bar", wantErr: true},
		{ref: "main-worktree/HEAD", target: "refs/heads/main", wantErr: false},
		{ref: "main-worktree/HEAD", target: "refs/tags/v1", wantErr: true},
	}

	for _, tt := range tests {
		err := name.ValidateSymbolicTarget(tt.ref, tt.target)
		if (err != nil) != tt.wantErr {
			t.Fatalf("ValidateSymbolicTarget(%q, %q) err=%v, wantErr=%v", tt.ref, tt.target, err, tt.wantErr)
		}
	}
}
