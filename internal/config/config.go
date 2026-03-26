package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	GHQ      GHQConfig      `toml:"ghq"`
	ROS      ROSConfig      `toml:"ros"`
	Defaults DefaultsConfig `toml:"defaults"`
	Alias    AliasConfig    `toml:"alias"`
}

type GHQConfig struct {
	Root string `toml:"root"`
}

type ROSConfig struct {
	Workspaces []WorkspaceDef `toml:"workspaces"`
}

type WorkspaceDef struct {
	Name      string `toml:"name"`
	Path      string `toml:"path"`
	SrcSubdir string `toml:"src_subdir"`
}

type DefaultsConfig struct {
	ROSWorkspace string `toml:"ros_workspace"`
}

type AliasConfig struct {
	Mode string `toml:"mode"`
}

// ConfigPath returns the path to the config file, respecting XDG_CONFIG_HOME.
func ConfigPath() string {
	if p := os.Getenv("RGW_CONFIG"); p != "" {
		return p
	}
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "rgw", "config.toml")
}

// Load reads config from the TOML file and overlays environment variables.
func Load() (*Config, error) {
	cfg := &Config{}
	cfg.setDefaults()

	path := ConfigPath()
	if _, err := os.Stat(path); err == nil {
		if _, err := toml.DecodeFile(path, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config %s: %w", path, err)
		}
	}

	cfg.applyEnv()
	cfg.expandPaths()

	return cfg, nil
}

// Save persists the config to the TOML file.
func (c *Config) Save() error {
	path := ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(c)
}

// ResolveWorkspace returns the active workspace definition.
// Precedence: flagWS > RGW_WS env > RGW_WS_PATH env > defaults.ros_workspace > first in list.
func (c *Config) ResolveWorkspace(flagWS string) (*WorkspaceDef, error) {
	name := flagWS
	if name == "" {
		name = os.Getenv("RGW_WS")
	}

	if name != "" {
		for i := range c.ROS.Workspaces {
			if c.ROS.Workspaces[i].Name == name {
				return &c.ROS.Workspaces[i], nil
			}
		}
		return nil, fmt.Errorf("workspace %q not found in config", name)
	}

	if p := os.Getenv("RGW_WS_PATH"); p != "" {
		return &WorkspaceDef{
			Name:      p,
			Path:      expandHome(p),
			SrcSubdir: "src",
		}, nil
	}

	if c.Defaults.ROSWorkspace != "" {
		for i := range c.ROS.Workspaces {
			if c.ROS.Workspaces[i].Name == c.Defaults.ROSWorkspace {
				return &c.ROS.Workspaces[i], nil
			}
		}
	}

	if len(c.ROS.Workspaces) > 0 {
		return &c.ROS.Workspaces[0], nil
	}

	return nil, fmt.Errorf("no workspace configured; use 'rgw ws add' to add one")
}

// ResolveGHQRoot returns the ghq root directory.
func (c *Config) ResolveGHQRoot() (string, error) {
	if p := os.Getenv("RGW_GHQ_ROOT"); p != "" {
		return expandHome(p), nil
	}
	if c.GHQ.Root != "" {
		return c.GHQ.Root, nil
	}
	out, err := exec.Command("ghq", "root").Output()
	if err != nil {
		return "", fmt.Errorf("failed to detect ghq root: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ResolveAliasMode returns the alias mode string.
func (c *Config) ResolveAliasMode() string {
	if m := os.Getenv("RGW_ALIAS_MODE"); m != "" {
		return m
	}
	if c.Alias.Mode != "" {
		return c.Alias.Mode
	}
	return "org_repo"
}

func (c *Config) setDefaults() {
	c.Alias.Mode = "org_repo"
}

func (c *Config) applyEnv() {
	if v := os.Getenv("RGW_GHQ_ROOT"); v != "" {
		c.GHQ.Root = expandHome(v)
	}
	if v := os.Getenv("RGW_ALIAS_MODE"); v != "" {
		c.Alias.Mode = v
	}
}

func (c *Config) expandPaths() {
	c.GHQ.Root = expandHome(c.GHQ.Root)
	for i := range c.ROS.Workspaces {
		c.ROS.Workspaces[i].Path = expandHome(c.ROS.Workspaces[i].Path)
		if c.ROS.Workspaces[i].SrcSubdir == "" {
			c.ROS.Workspaces[i].SrcSubdir = "src"
		}
	}
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
