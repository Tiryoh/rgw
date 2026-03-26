package ghq

import (
	"os"
	"path/filepath"
	"testing"
)

func setupGHQTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	// Create github.com/Tiryoh/my_robot
	os.MkdirAll(filepath.Join(root, "github.com", "Tiryoh", "my_robot"), 0o755)
	// Create github.com/OtherOrg/my_robot
	os.MkdirAll(filepath.Join(root, "github.com", "OtherOrg", "my_robot"), 0o755)
	// Create github.com/Tiryoh/nav_pkg
	os.MkdirAll(filepath.Join(root, "github.com", "Tiryoh", "nav_pkg"), 0o755)
	return root
}

func TestParseRepoArgFullyQualified(t *testing.T) {
	root := setupGHQTree(t)
	info, path, err := ParseRepoArg(root, "github.com/Tiryoh/my_robot")
	if err != nil {
		t.Fatal(err)
	}
	if info.Host != "github.com" || info.Org != "Tiryoh" || info.Repo != "my_robot" {
		t.Errorf("info = %+v", info)
	}
	want := filepath.Join(root, "github.com", "Tiryoh", "my_robot")
	if path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
}

func TestParseRepoArgOrgRepo(t *testing.T) {
	root := setupGHQTree(t)
	info, _, err := ParseRepoArg(root, "Tiryoh/nav_pkg")
	if err != nil {
		t.Fatal(err)
	}
	if info.Org != "Tiryoh" || info.Repo != "nav_pkg" {
		t.Errorf("info = %+v", info)
	}
}

func TestParseRepoArgShortAmbiguous(t *testing.T) {
	root := setupGHQTree(t)
	_, _, err := ParseRepoArg(root, "my_robot")
	if err == nil {
		t.Fatal("expected error for ambiguous repo name")
	}
}

func TestParseRepoArgShortUnique(t *testing.T) {
	root := setupGHQTree(t)
	info, _, err := ParseRepoArg(root, "nav_pkg")
	if err != nil {
		t.Fatal(err)
	}
	if info.Repo != "nav_pkg" {
		t.Errorf("info.Repo = %q, want %q", info.Repo, "nav_pkg")
	}
}

func TestParseRepoArgNotFound(t *testing.T) {
	root := setupGHQTree(t)
	_, _, err := ParseRepoArg(root, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent repo")
	}
}
