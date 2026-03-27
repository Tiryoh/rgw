package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Tiryoh/rgw/internal/config"
	"github.com/Tiryoh/rgw/internal/validate"
)

// List returns all configured workspaces.
func List(cfg *config.Config) []config.WorkspaceDef {
	return cfg.ROS.Workspaces
}

// Add appends a new workspace to config and saves.
func Add(cfg *config.Config, name, path string) error {
	if err := validate.WorkspaceName(name); err != nil {
		return fmt.Errorf("invalid workspace name %q: %w", name, err)
	}
	if err := validate.NoControlChars(path); err != nil {
		return fmt.Errorf("invalid workspace path %q: %w", path, err)
	}
	for _, ws := range cfg.ROS.Workspaces {
		if ws.Name == name {
			return fmt.Errorf("workspace %q already exists", name)
		}
	}
	cfg.ROS.Workspaces = append(cfg.ROS.Workspaces, config.WorkspaceDef{
		Name:      name,
		Path:      path,
		SrcSubdir: "src",
	})
	if len(cfg.ROS.Workspaces) == 1 {
		cfg.Defaults.ROSWorkspace = name
	}
	return cfg.Save()
}

// Use sets the default workspace and saves.
func Use(cfg *config.Config, name string) error {
	found := false
	for _, ws := range cfg.ROS.Workspaces {
		if ws.Name == name {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("workspace %q not found", name)
	}
	cfg.Defaults.ROSWorkspace = name
	return cfg.Save()
}

// Current returns the currently active workspace definition.
func Current(cfg *config.Config, flagWS string) (*config.WorkspaceDef, error) {
	return cfg.ResolveWorkspace(flagWS)
}

// ValidatePath checks that the ws path exists and has a src/ subdirectory.
func ValidatePath(wsDef *config.WorkspaceDef) error {
	srcDir := filepath.Join(wsDef.Path, wsDef.SrcSubdir)
	info, err := os.Stat(srcDir)
	if err != nil {
		return fmt.Errorf("workspace src directory does not exist: %s", srcDir)
	}
	if !info.IsDir() {
		return fmt.Errorf("workspace src path is not a directory: %s", srcDir)
	}
	return nil
}
