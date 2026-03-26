package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Tiryoh/rgw/internal/config"
)

func TestCheckBuildArtifacts(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "src"), 0o755)
	os.MkdirAll(filepath.Join(dir, "build"), 0o755)
	os.MkdirAll(filepath.Join(dir, "install"), 0o755)

	ws := &config.WorkspaceDef{Name: "test", Path: dir, SrcSubdir: "src"}
	results := CheckBuildArtifacts(ws, "test")
	if len(results) != 2 {
		t.Errorf("got %d results, want 2 (build + install)", len(results))
	}
	for _, r := range results {
		if r.Severity != SeverityWarn {
			t.Errorf("expected WARN, got %v", r.Severity)
		}
	}
}

func TestCheckBrokenSymlinks(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	os.MkdirAll(srcDir, 0o755)

	// Create a broken symlink
	os.Symlink("/nonexistent/target", filepath.Join(srcDir, "broken"))
	// Create a valid symlink
	target := t.TempDir()
	os.Symlink(target, filepath.Join(srcDir, "valid"))

	ws := &config.WorkspaceDef{Name: "test", Path: dir, SrcSubdir: "src"}
	results := CheckBrokenSymlinks(ws, "test")
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Severity != SeverityError {
		t.Errorf("expected ERR, got %v", results[0].Severity)
	}
}

func TestRunAllNoWorkspaces(t *testing.T) {
	t.Setenv("RGW_CONFIG", "/nonexistent")
	cfg := &config.Config{}
	// Set ghq root to avoid ghq command execution
	t.Setenv("RGW_GHQ_ROOT", t.TempDir())

	results := RunAll(cfg)
	var found bool
	for _, r := range results {
		if r.Name == "workspaces" && r.Severity == SeverityWarn {
			found = true
		}
	}
	if !found {
		t.Error("expected warning about no workspaces configured")
	}
}
