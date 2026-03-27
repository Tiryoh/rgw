package symlink

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Tiryoh/rgw/internal/config"
	"github.com/Tiryoh/rgw/internal/validate"
)

// Set creates or replaces a symlink at <ws>/src/<alias> -> worktreePath.
// Equivalent to `ln -sfn`.
func Set(wsDef *config.WorkspaceDef, alias string, worktreePath string) error {
	if _, err := os.Stat(worktreePath); err != nil {
		return fmt.Errorf("target path does not exist: %s", worktreePath)
	}

	srcDir := filepath.Join(wsDef.Path, wsDef.SrcSubdir)
	if _, err := os.Stat(srcDir); err != nil {
		return fmt.Errorf("workspace src directory does not exist: %s", srcDir)
	}

	linkPath, err := validate.SafePath(srcDir, alias)
	if err != nil {
		return fmt.Errorf("unsafe alias %q: %w", alias, err)
	}

	// Check existing entry at linkPath
	linfo, err := os.Lstat(linkPath)
	if err == nil {
		if linfo.Mode()&os.ModeSymlink != 0 {
			// Existing symlink — remove it
			if err := os.Remove(linkPath); err != nil {
				return fmt.Errorf("failed to remove existing symlink: %w", err)
			}
		} else {
			// Real file/directory — refuse to overwrite
			return fmt.Errorf("path %s exists and is not a symlink; refusing to overwrite", linkPath)
		}
	}

	if err := os.Symlink(worktreePath, linkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// Unset removes the symlink at <ws>/src/<alias>.
// Returns error if the path is not a symlink (safety: never remove real directories).
func Unset(wsDef *config.WorkspaceDef, alias string) error {
	srcDir := filepath.Join(wsDef.Path, wsDef.SrcSubdir)
	linkPath, err := validate.SafePath(srcDir, alias)
	if err != nil {
		return fmt.Errorf("unsafe alias %q: %w", alias, err)
	}

	linfo, err := os.Lstat(linkPath)
	if err != nil {
		return fmt.Errorf("symlink does not exist: %s", linkPath)
	}
	if linfo.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("path %s is not a symlink; refusing to remove", linkPath)
	}

	return os.Remove(linkPath)
}
