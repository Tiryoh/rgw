package symlink

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		// Normalize relative symlink targets to absolute paths.
		if !filepath.IsAbs(target) {
			target = filepath.Join(srcDir, target)
		}
		target = filepath.Clean(target)

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

// FindByAlias returns the Link with the given alias, or an error if not found.
func FindByAlias(wsDef *config.WorkspaceDef, alias string) (*Link, error) {
	links, err := Status(wsDef)
	if err != nil {
		return nil, err
	}
	for i := range links {
		if links[i].Alias == alias {
			return &links[i], nil
		}
	}
	return nil, fmt.Errorf("alias %q not found; run 'rgw link status' to see existing links", alias)
}

// Repair removes broken symlinks in <ws>/src/.
func Repair(wsDef *config.WorkspaceDef) (removed int, err error) {
	links, err := Status(wsDef)
	if err != nil {
		return 0, err
	}
	var errs []string
	for _, link := range links {
		if !link.Orphaned {
			continue
		}
		// Re-verify the path is still a symlink before removing (TOCTOU guard).
		linfo, statErr := os.Lstat(link.LinkPath)
		if statErr != nil {
			// Already gone — count as success.
			removed++
			continue
		}
		if linfo.Mode()&os.ModeSymlink == 0 {
			errs = append(errs, fmt.Sprintf("%s is no longer a symlink", link.LinkPath))
			continue
		}
		if removeErr := os.Remove(link.LinkPath); removeErr != nil {
			errs = append(errs, removeErr.Error())
			continue
		}
		removed++
	}
	if len(errs) > 0 {
		return removed, fmt.Errorf("repair: %d error(s): %s",
			len(errs), strings.Join(errs, "; "))
	}
	return removed, nil
}
