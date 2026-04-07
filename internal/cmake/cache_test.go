package cmake

import (
	"os"
	"path/filepath"
	"testing"
)

func writeCMakeCache(t *testing.T, buildDir, pkg, homeDir string) {
	t.Helper()
	dir := filepath.Join(buildDir, pkg)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "# CMakeCache\n" +
		"CMAKE_CACHEFILE_DIR:INTERNAL=" + dir + "\n" +
		"CMAKE_HOME_DIRECTORY:INTERNAL=" + homeDir + "\n"
	if err := os.WriteFile(filepath.Join(dir, "CMakeCache.txt"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCheckStaleCache_DetectsStale(t *testing.T) {
	wsPath := t.TempDir()
	buildDir := filepath.Join(wsPath, "build")

	oldTarget := "/home/user/worktree/org/repo/feature-old"
	newTarget := "/home/user/worktree/org/repo/feature-new"

	// Stale: points to old worktree
	writeCMakeCache(t, buildDir, "pkg_a", oldTarget+"/pkg_a")
	// Fresh: points to new worktree
	writeCMakeCache(t, buildDir, "pkg_b", newTarget+"/pkg_b")
	// Unrelated: points to completely different path
	writeCMakeCache(t, buildDir, "pkg_c", "/opt/ros/humble/share/pkg_c")

	stale, err := CheckStaleCache(wsPath, oldTarget)
	if err != nil {
		t.Fatal(err)
	}
	if len(stale) != 1 {
		t.Fatalf("expected 1 stale package, got %d", len(stale))
	}
	if stale[0].Package != "pkg_a" {
		t.Errorf("expected stale package pkg_a, got %s", stale[0].Package)
	}
}

func TestCheckStaleCache_NoBuildDir(t *testing.T) {
	wsPath := t.TempDir()
	stale, err := CheckStaleCache(wsPath, "/some/path")
	if err != nil {
		t.Fatal(err)
	}
	if len(stale) != 0 {
		t.Errorf("expected 0 stale packages, got %d", len(stale))
	}
}

func TestCheckStaleCache_NoCacheFile(t *testing.T) {
	wsPath := t.TempDir()
	buildDir := filepath.Join(wsPath, "build", "pkg_no_cache")
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		t.Fatal(err)
	}

	stale, err := CheckStaleCache(wsPath, "/some/path")
	if err != nil {
		t.Fatal(err)
	}
	if len(stale) != 0 {
		t.Errorf("expected 0 stale packages, got %d", len(stale))
	}
}

func TestFormatWarning_FewPackages(t *testing.T) {
	stale := []StalePackage{
		{Package: "pkg_a", CachedDir: "/old/path/pkg_a"},
	}
	msg := FormatWarning("/home/user/ros2_ws", stale)
	if msg == "" {
		t.Fatal("expected non-empty warning")
	}
	if !contains(msg, "pkg_a") {
		t.Error("warning should mention stale package name")
	}
	if !contains(msg, "rm -rf") {
		t.Error("warning should include rm command")
	}
}

func TestFormatWarning_ManyPackages(t *testing.T) {
	stale := make([]StalePackage, 6)
	for i := range stale {
		stale[i] = StalePackage{Package: "pkg_" + string(rune('a'+i)), CachedDir: "/old/path"}
	}
	msg := FormatWarning("/ws", stale)
	// Should use brace expansion format for >5 packages
	if !contains(msg, "{") {
		t.Error("warning should use brace expansion for many packages")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
