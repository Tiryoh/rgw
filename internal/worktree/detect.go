package worktree

import (
	"fmt"
	"os/exec"
	"strings"
)

// ListForRepo runs `git worktree list --porcelain` for the given repo path
// and returns parsed worktree entries.
func ListForRepo(repoPath string) ([]Worktree, error) {
	cmd := exec.Command("git", "-C", repoPath, "worktree", "list", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list failed for %s: %w", repoPath, err)
	}
	return ParsePorcelain(out)
}

// ParsePorcelain parses the raw output of `git worktree list --porcelain`.
//
// Format:
//
//	worktree <path>
//	HEAD <sha>
//	branch refs/heads/<name>
//	<blank line>
func ParsePorcelain(data []byte) ([]Worktree, error) {
	text := strings.TrimSpace(string(data))
	if text == "" {
		return nil, nil
	}

	// Split into stanzas separated by blank lines.
	stanzas := splitStanzas(text)
	result := make([]Worktree, 0, len(stanzas))

	for _, stanza := range stanzas {
		wt, err := parseStanza(stanza)
		if err != nil {
			return nil, err
		}
		result = append(result, wt)
	}

	return result, nil
}

// IsDirty checks if a worktree has uncommitted changes.
func IsDirty(worktreePath string) (bool, error) {
	cmd := exec.Command("git", "-C", worktreePath, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status failed for %s: %w", worktreePath, err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}

func splitStanzas(text string) [][]string {
	var stanzas [][]string
	var current []string
	for _, line := range strings.Split(text, "\n") {
		if line == "" {
			if len(current) > 0 {
				stanzas = append(stanzas, current)
				current = nil
			}
			continue
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		stanzas = append(stanzas, current)
	}
	return stanzas
}

func parseStanza(lines []string) (Worktree, error) {
	var wt Worktree
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "worktree "):
			wt.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "HEAD "):
			wt.HEAD = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			wt.Branch = strings.TrimPrefix(ref, "refs/heads/")
		case line == "bare":
			wt.Bare = true
		case line == "detached":
			wt.Detached = true
		case line == "prunable":
			wt.Prunable = true
		}
	}
	if wt.Path == "" {
		return wt, fmt.Errorf("worktree stanza missing path")
	}
	return wt, nil
}
