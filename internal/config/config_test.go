package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	content := `
[ghq]
root = "/home/user/ghq"

[ros]
[[ros.workspaces]]
name = "default"
path = "/home/user/ros2_ws"
src_subdir = "src"

[[ros.workspaces]]
name = "nav"
path = "/home/user/nav_ws"

[defaults]
ros_workspace = "default"

[alias]
mode = "org_repo"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("RGW_CONFIG", cfgPath)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.GHQ.Root != "/home/user/ghq" {
		t.Errorf("GHQ.Root = %q, want %q", cfg.GHQ.Root, "/home/user/ghq")
	}
	if len(cfg.ROS.Workspaces) != 2 {
		t.Fatalf("len(Workspaces) = %d, want 2", len(cfg.ROS.Workspaces))
	}
	if cfg.ROS.Workspaces[0].Name != "default" {
		t.Errorf("Workspaces[0].Name = %q, want %q", cfg.ROS.Workspaces[0].Name, "default")
	}
	if cfg.ROS.Workspaces[1].SrcSubdir != "src" {
		t.Errorf("Workspaces[1].SrcSubdir = %q, want %q (default)", cfg.ROS.Workspaces[1].SrcSubdir, "src")
	}
	if cfg.Defaults.ROSWorkspace != "default" {
		t.Errorf("Defaults.ROSWorkspace = %q, want %q", cfg.Defaults.ROSWorkspace, "default")
	}
	if cfg.Alias.Mode != "org_repo" {
		t.Errorf("Alias.Mode = %q, want %q", cfg.Alias.Mode, "org_repo")
	}
}

func TestLoadMissingFileUsesDefaults(t *testing.T) {
	t.Setenv("RGW_CONFIG", "/nonexistent/path/config.toml")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Alias.Mode != "repo" {
		t.Errorf("default Alias.Mode = %q, want %q", cfg.Alias.Mode, "repo")
	}
}

func TestEnvOverlay(t *testing.T) {
	t.Setenv("RGW_CONFIG", "/nonexistent/path/config.toml")
	t.Setenv("RGW_GHQ_ROOT", "/env/ghq")
	t.Setenv("RGW_ALIAS_MODE", "host_org_repo")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.GHQ.Root != "/env/ghq" {
		t.Errorf("GHQ.Root = %q, want %q", cfg.GHQ.Root, "/env/ghq")
	}
	if cfg.ResolveAliasMode() != "host_org_repo" {
		t.Errorf("ResolveAliasMode() = %q, want %q", cfg.ResolveAliasMode(), "host_org_repo")
	}
}

func TestResolveWorkspace(t *testing.T) {
	cfg := &Config{
		ROS: ROSConfig{
			Workspaces: []WorkspaceDef{
				{Name: "ws1", Path: "/ws1", SrcSubdir: "src"},
				{Name: "ws2", Path: "/ws2", SrcSubdir: "src"},
			},
		},
		Defaults: DefaultsConfig{ROSWorkspace: "ws1"},
	}

	// flag takes precedence
	ws, err := cfg.ResolveWorkspace("ws2")
	if err != nil {
		t.Fatal(err)
	}
	if ws.Name != "ws2" {
		t.Errorf("Name = %q, want %q", ws.Name, "ws2")
	}

	// falls back to default
	ws, err = cfg.ResolveWorkspace("")
	if err != nil {
		t.Fatal(err)
	}
	if ws.Name != "ws1" {
		t.Errorf("Name = %q, want %q", ws.Name, "ws1")
	}

	// not found
	_, err = cfg.ResolveWorkspace("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent workspace")
	}
}

func TestResolveWorkspaceFromEnv(t *testing.T) {
	cfg := &Config{
		ROS: ROSConfig{
			Workspaces: []WorkspaceDef{
				{Name: "ws1", Path: "/ws1", SrcSubdir: "src"},
			},
		},
	}

	t.Setenv("RGW_WS", "ws1")
	ws, err := cfg.ResolveWorkspace("")
	if err != nil {
		t.Fatal(err)
	}
	if ws.Name != "ws1" {
		t.Errorf("Name = %q, want %q", ws.Name, "ws1")
	}
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := expandHome("~/foo")
	want := filepath.Join(home, "foo")
	if got != want {
		t.Errorf("expandHome(~/foo) = %q, want %q", got, want)
	}

	got = expandHome("/absolute/path")
	if got != "/absolute/path" {
		t.Errorf("expandHome(/absolute/path) = %q, want %q", got, "/absolute/path")
	}
}
