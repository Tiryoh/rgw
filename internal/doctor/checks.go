package doctor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Tiryoh/rgw/internal/config"
	"github.com/Tiryoh/rgw/internal/symlink"
	"github.com/Tiryoh/rgw/internal/worktree"
)

// Severity indicates the severity of a check result.
type Severity int

const (
	SeverityOK Severity = iota
	SeverityWarn
	SeverityError
)

func (s Severity) String() string {
	switch s {
	case SeverityOK:
		return "OK"
	case SeverityWarn:
		return "WARN"
	case SeverityError:
		return "ERR"
	default:
		return "?"
	}
}

// CheckResult represents the outcome of a single health check.
type CheckResult struct {
	Name     string
	Severity Severity
	Message  string
}

// RunAll runs all health checks and returns results.
func RunAll(cfg *config.Config) []CheckResult {
	var results []CheckResult

	// Check ghq root
	ghqRoot, err := cfg.ResolveGHQRoot()
	if err != nil {
		results = append(results, CheckResult{
			Name:     "ghq_root",
			Severity: SeverityError,
			Message:  fmt.Sprintf("Cannot resolve ghq root: %v", err),
		})
	} else {
		results = append(results, CheckResult{
			Name:     "ghq_root",
			Severity: SeverityOK,
			Message:  fmt.Sprintf("ghq root: %s", ghqRoot),
		})
	}

	// Check each workspace
	for _, ws := range cfg.ROS.Workspaces {
		results = append(results, checkWorkspace(&ws)...)
	}

	if len(cfg.ROS.Workspaces) == 0 {
		results = append(results, CheckResult{
			Name:     "workspaces",
			Severity: SeverityWarn,
			Message:  "No workspaces configured",
		})
	}

	return results
}

func checkWorkspace(ws *config.WorkspaceDef) []CheckResult {
	var results []CheckResult
	prefix := ws.Name

	// Check workspace path exists
	if _, err := os.Stat(ws.Path); err != nil {
		results = append(results, CheckResult{
			Name:     prefix + "/path",
			Severity: SeverityError,
			Message:  fmt.Sprintf("Workspace path does not exist: %s", ws.Path),
		})
		return results
	}

	// Check src subdir exists
	srcDir := filepath.Join(ws.Path, ws.SrcSubdir)
	if _, err := os.Stat(srcDir); err != nil {
		results = append(results, CheckResult{
			Name:     prefix + "/src",
			Severity: SeverityError,
			Message:  fmt.Sprintf("src directory does not exist: %s", srcDir),
		})
		return results
	}

	results = append(results, CheckResult{
		Name:     prefix + "/path",
		Severity: SeverityOK,
		Message:  fmt.Sprintf("Workspace: %s", ws.Path),
	})

	// Check build artifacts
	results = append(results, CheckBuildArtifacts(ws, prefix)...)

	// Check broken symlinks
	results = append(results, CheckBrokenSymlinks(ws, prefix)...)

	// Check dirty worktrees
	results = append(results, CheckDirtyWorktrees(ws, prefix)...)

	return results
}

// CheckBuildArtifacts checks for build/install/log directories in the workspace.
func CheckBuildArtifacts(ws *config.WorkspaceDef, prefix string) []CheckResult {
	var results []CheckResult
	for _, dir := range []string{"build", "install", "log"} {
		p := filepath.Join(ws.Path, dir)
		if _, err := os.Stat(p); err == nil {
			results = append(results, CheckResult{
				Name:     prefix + "/" + dir,
				Severity: SeverityWarn,
				Message:  fmt.Sprintf("%s/ directory exists (stale build artifacts?)", dir),
			})
		}
	}
	return results
}

// CheckBrokenSymlinks checks for dangling symlinks in src/.
func CheckBrokenSymlinks(ws *config.WorkspaceDef, prefix string) []CheckResult {
	links, err := symlink.Status(ws)
	if err != nil {
		return []CheckResult{{
			Name:     prefix + "/symlinks",
			Severity: SeverityError,
			Message:  fmt.Sprintf("Failed to scan symlinks: %v", err),
		}}
	}

	var broken int
	for _, l := range links {
		if l.Orphaned {
			broken++
		}
	}
	if broken > 0 {
		return []CheckResult{{
			Name:     prefix + "/symlinks",
			Severity: SeverityError,
			Message:  fmt.Sprintf("%d broken symlink(s) in %s/%s", broken, ws.Path, ws.SrcSubdir),
		}}
	}
	return nil
}

// CheckDirtyWorktrees checks if linked worktrees have uncommitted changes.
func CheckDirtyWorktrees(ws *config.WorkspaceDef, prefix string) []CheckResult {
	links, err := symlink.Status(ws)
	if err != nil {
		return nil
	}

	var results []CheckResult
	for _, l := range links {
		if !l.Valid {
			continue
		}
		dirty, err := worktree.IsDirty(l.TargetPath)
		if err != nil {
			continue
		}
		if dirty {
			results = append(results, CheckResult{
				Name:     prefix + "/dirty",
				Severity: SeverityWarn,
				Message:  fmt.Sprintf("Linked worktree has uncommitted changes: %s", l.TargetPath),
			})
		}
	}
	return results
}
