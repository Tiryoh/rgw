package worktree

// Worktree represents a single git worktree entry.
type Worktree struct {
	Path     string // absolute filesystem path
	Branch   string // branch name without refs/heads/ prefix
	HEAD     string // commit SHA
	Bare     bool
	Detached bool
	Prunable bool
}
