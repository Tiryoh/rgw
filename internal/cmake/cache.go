package cmake

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// StalePackage represents a build directory whose CMake cache
// still references an old worktree source path.
type StalePackage struct {
	Package  string // package directory name under build/
	CachedDir string // CMAKE_HOME_DIRECTORY value found in cache
}

// CheckStaleCache scans <wsPath>/build/*/CMakeCache.txt for entries
// whose CMAKE_HOME_DIRECTORY contains oldTargetPath (the previous
// worktree symlink target). Returns the list of stale packages.
func CheckStaleCache(wsPath string, oldTargetPath string) ([]StalePackage, error) {
	buildDir := filepath.Join(wsPath, "build")
	entries, err := os.ReadDir(buildDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var stale []StalePackage
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		cacheFile := filepath.Join(buildDir, e.Name(), "CMakeCache.txt")
		cachedDir, err := readCMakeHomeDir(cacheFile)
		if err != nil || cachedDir == "" {
			continue
		}
		if strings.HasPrefix(cachedDir, oldTargetPath+"/") || cachedDir == oldTargetPath {
			stale = append(stale, StalePackage{
				Package:   e.Name(),
				CachedDir: cachedDir,
			})
		}
	}
	return stale, nil
}

// readCMakeHomeDir extracts the CMAKE_HOME_DIRECTORY value from a
// CMakeCache.txt file. Returns "" if the file doesn't exist or the
// key is not found.
func readCMakeHomeDir(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	const prefix = "CMAKE_HOME_DIRECTORY:INTERNAL="
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, prefix) {
			return line[len(prefix):], nil
		}
	}
	return "", scanner.Err()
}

// FormatWarning produces a human-readable warning message for stale
// CMake cache entries, including suggested cleanup commands.
func FormatWarning(wsPath string, stale []StalePackage) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\nWarning: %d package(s) have stale CMake cache entries:\n", len(stale))
	for _, s := range stale {
		fmt.Fprintf(&b, "  - %s (cached: %s)\n", s.Package, s.CachedDir)
	}
	buildDir := filepath.Join(wsPath, "build")
	b.WriteString("\nTo fix, remove the stale build directories:\n")
	if len(stale) <= 5 {
		for _, s := range stale {
			fmt.Fprintf(&b, "  rm -rf %s\n", filepath.Join(buildDir, s.Package))
		}
	} else {
		fmt.Fprintf(&b, "  rm -rf %s/{", buildDir)
		names := make([]string, len(stale))
		for i, s := range stale {
			names[i] = s.Package
		}
		b.WriteString(strings.Join(names, ","))
		b.WriteString("}\n")
	}
	fmt.Fprintf(&b, "\nOr clean the entire workspace build:\n  rm -rf %s %s %s\n",
		buildDir,
		filepath.Join(wsPath, "install"),
		filepath.Join(wsPath, "log"),
	)
	return b.String()
}
