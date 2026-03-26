package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Tiryoh/rgw/internal/config"
)

func TestAddAndList(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	t.Setenv("RGW_CONFIG", cfgPath)

	cfg := &config.Config{}

	if err := Add(cfg, "ws1", "/path/to/ws1"); err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	wsList := List(cfg)
	if len(wsList) != 1 {
		t.Fatalf("len(List()) = %d, want 1", len(wsList))
	}
	if wsList[0].Name != "ws1" {
		t.Errorf("Name = %q, want %q", wsList[0].Name, "ws1")
	}

	// first workspace should become default
	if cfg.Defaults.ROSWorkspace != "ws1" {
		t.Errorf("Defaults.ROSWorkspace = %q, want %q", cfg.Defaults.ROSWorkspace, "ws1")
	}

	// duplicate
	if err := Add(cfg, "ws1", "/other"); err == nil {
		t.Error("expected error for duplicate workspace name")
	}
}

func TestUse(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	t.Setenv("RGW_CONFIG", cfgPath)

	cfg := &config.Config{}
	Add(cfg, "ws1", "/ws1")
	Add(cfg, "ws2", "/ws2")

	if err := Use(cfg, "ws2"); err != nil {
		t.Fatalf("Use() error: %v", err)
	}
	if cfg.Defaults.ROSWorkspace != "ws2" {
		t.Errorf("Defaults.ROSWorkspace = %q, want %q", cfg.Defaults.ROSWorkspace, "ws2")
	}

	if err := Use(cfg, "nonexistent"); err == nil {
		t.Error("expected error for nonexistent workspace")
	}
}

func TestValidatePath(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	os.MkdirAll(srcDir, 0o755)

	ws := &config.WorkspaceDef{Path: dir, SrcSubdir: "src"}
	if err := ValidatePath(ws); err != nil {
		t.Errorf("ValidatePath() error: %v", err)
	}

	ws2 := &config.WorkspaceDef{Path: "/nonexistent", SrcSubdir: "src"}
	if err := ValidatePath(ws2); err == nil {
		t.Error("expected error for nonexistent path")
	}
}
