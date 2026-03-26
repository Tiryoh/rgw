package symlink

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Tiryoh/rgw/internal/config"
)

func setupWS(t *testing.T) (*config.WorkspaceDef, string) {
	t.Helper()
	wsDir := t.TempDir()
	srcDir := filepath.Join(wsDir, "src")
	os.MkdirAll(srcDir, 0o755)

	targetDir := t.TempDir() // simulates a worktree path
	return &config.WorkspaceDef{
		Name:      "test",
		Path:      wsDir,
		SrcSubdir: "src",
	}, targetDir
}

func TestSetAndUnset(t *testing.T) {
	wsDef, target := setupWS(t)

	// Set
	if err := Set(wsDef, "test__pkg", target); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	linkPath := filepath.Join(wsDef.Path, "src", "test__pkg")
	got, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("Readlink error: %v", err)
	}
	if got != target {
		t.Errorf("symlink target = %q, want %q", got, target)
	}

	// Unset
	if err := Unset(wsDef, "test__pkg"); err != nil {
		t.Fatalf("Unset() error: %v", err)
	}
	if _, err := os.Lstat(linkPath); !os.IsNotExist(err) {
		t.Error("symlink should be removed after Unset")
	}
}

func TestSetReplacesExistingSymlink(t *testing.T) {
	wsDef, target1 := setupWS(t)
	target2 := t.TempDir()

	Set(wsDef, "pkg", target1)
	if err := Set(wsDef, "pkg", target2); err != nil {
		t.Fatalf("Set() replace error: %v", err)
	}

	linkPath := filepath.Join(wsDef.Path, "src", "pkg")
	got, _ := os.Readlink(linkPath)
	if got != target2 {
		t.Errorf("symlink target = %q, want %q", got, target2)
	}
}

func TestSetRefusesRealDirectory(t *testing.T) {
	wsDef, _ := setupWS(t)
	realDir := filepath.Join(wsDef.Path, "src", "real_pkg")
	os.MkdirAll(realDir, 0o755)

	target := t.TempDir()
	err := Set(wsDef, "real_pkg", target)
	if err == nil {
		t.Fatal("expected error when alias path is a real directory")
	}
}

func TestUnsetRefusesNonSymlink(t *testing.T) {
	wsDef, _ := setupWS(t)
	realDir := filepath.Join(wsDef.Path, "src", "real_pkg")
	os.MkdirAll(realDir, 0o755)

	err := Unset(wsDef, "real_pkg")
	if err == nil {
		t.Fatal("expected error when trying to unset a real directory")
	}
}

func TestStatus(t *testing.T) {
	wsDef, target := setupWS(t)
	Set(wsDef, "pkg1", target)

	// Create an orphaned symlink
	orphanLink := filepath.Join(wsDef.Path, "src", "orphan")
	os.Symlink("/nonexistent/path", orphanLink)

	links, err := Status(wsDef)
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("got %d links, want 2", len(links))
	}

	var validCount, orphanCount int
	for _, l := range links {
		if l.Valid {
			validCount++
		}
		if l.Orphaned {
			orphanCount++
		}
	}
	if validCount != 1 {
		t.Errorf("valid count = %d, want 1", validCount)
	}
	if orphanCount != 1 {
		t.Errorf("orphan count = %d, want 1", orphanCount)
	}
}

func TestRepair(t *testing.T) {
	wsDef, target := setupWS(t)
	Set(wsDef, "good", target)
	os.Symlink("/nonexistent", filepath.Join(wsDef.Path, "src", "broken"))

	removed, err := Repair(wsDef)
	if err != nil {
		t.Fatalf("Repair() error: %v", err)
	}
	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}

	links, _ := Status(wsDef)
	if len(links) != 1 {
		t.Errorf("links after repair = %d, want 1", len(links))
	}
}
