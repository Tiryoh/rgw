package worktree

import (
	"testing"
)

func TestParsePorcelain(t *testing.T) {
	input := []byte(`worktree /home/user/ghq/github.com/org/repo
HEAD abc123def456
branch refs/heads/main

worktree /home/user/ghq/github.com/org/repo-feat
HEAD def456abc789
branch refs/heads/feature/foo

`)
	wts, err := ParsePorcelain(input)
	if err != nil {
		t.Fatalf("ParsePorcelain() error: %v", err)
	}
	if len(wts) != 2 {
		t.Fatalf("got %d worktrees, want 2", len(wts))
	}

	if wts[0].Path != "/home/user/ghq/github.com/org/repo" {
		t.Errorf("wts[0].Path = %q", wts[0].Path)
	}
	if wts[0].Branch != "main" {
		t.Errorf("wts[0].Branch = %q, want %q", wts[0].Branch, "main")
	}
	if wts[0].HEAD != "abc123def456" {
		t.Errorf("wts[0].HEAD = %q", wts[0].HEAD)
	}

	if wts[1].Branch != "feature/foo" {
		t.Errorf("wts[1].Branch = %q, want %q", wts[1].Branch, "feature/foo")
	}
}

func TestParsePorcelainDetached(t *testing.T) {
	input := []byte(`worktree /tmp/repo
HEAD abc123
detached

`)
	wts, err := ParsePorcelain(input)
	if err != nil {
		t.Fatalf("ParsePorcelain() error: %v", err)
	}
	if len(wts) != 1 {
		t.Fatalf("got %d worktrees, want 1", len(wts))
	}
	if !wts[0].Detached {
		t.Error("expected Detached = true")
	}
	if wts[0].Branch != "" {
		t.Errorf("Branch = %q, want empty", wts[0].Branch)
	}
}

func TestParsePorcelainBare(t *testing.T) {
	input := []byte(`worktree /tmp/bare.git
HEAD abc123
branch refs/heads/main
bare

`)
	wts, err := ParsePorcelain(input)
	if err != nil {
		t.Fatalf("ParsePorcelain() error: %v", err)
	}
	if !wts[0].Bare {
		t.Error("expected Bare = true")
	}
}

func TestParsePorcelainEmpty(t *testing.T) {
	wts, err := ParsePorcelain([]byte(""))
	if err != nil {
		t.Fatalf("ParsePorcelain() error: %v", err)
	}
	if len(wts) != 0 {
		t.Errorf("got %d worktrees, want 0", len(wts))
	}
}
