package symlink

import (
	"os"
	"path/filepath"

	"github.com/Tiryoh/rgw/internal/config"
)

// Link represents a symlink from a ROS workspace src/ to a worktree.
type Link struct {
	Alias      string `json:"alias"`
	LinkPath   string `json:"link_path"`
	TargetPath string `json:"target_path"`
	Valid      bool   `json:"valid"`
	Orphaned   bool   `json:"orphaned"`
	Branch     string `json:"branch,omitempty"`
}

// Status scans <ws>/src/ for all symlinks and returns their state.
func Status(wsDef *config.WorkspaceDef) ([]Link, error) {
	srcDir := filepath.Join(wsDef.Path, wsDef.SrcSubdir)
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return nil, err
	}

	var links []Link
	for _, entry := range entries {
		fullPath := filepath.Join(srcDir, entry.Name())
		linfo, err := os.Lstat(fullPath)
		if err != nil {
			continue
		}
		if linfo.Mode()&os.ModeSymlink == 0 {
			continue
		}

		target, err := os.Readlink(fullPath)
		if err != nil {
			continue
		}

		_, statErr := os.Stat(fullPath)
		valid := statErr == nil

		links = append(links, Link{
			Alias:      entry.Name(),
			LinkPath:   fullPath,
			TargetPath: target,
			Valid:      valid,
			Orphaned:   !valid,
		})
	}

	return links, nil
}

// Repair removes broken symlinks in <ws>/src/.
func Repair(wsDef *config.WorkspaceDef) (removed int, err error) {
	links, err := Status(wsDef)
	if err != nil {
		return 0, err
	}
	for _, link := range links {
		if link.Orphaned {
			if err := os.Remove(link.LinkPath); err != nil {
				return removed, err
			}
			removed++
		}
	}
	return removed, nil
}
