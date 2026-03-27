package worktree

// Worktree represents a single git worktree entry.
type Worktree struct {
	Path     string `json:"path"`     // absolute filesystem path
	Branch   string `json:"branch"`   // branch name without refs/heads/ prefix
	HEAD     string `json:"head"`     // commit SHA
	Bare     bool   `json:"bare"`     // bare repository
	Detached bool   `json:"detached"` // detached HEAD
	Prunable bool   `json:"prunable"` // prunable worktree
}
