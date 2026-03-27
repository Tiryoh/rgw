package validate

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// SafePath ensures that child, when joined to base, resolves to a path
// within base. Returns the cleaned joined path or an error.
func SafePath(base, child string) (string, error) {
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", fmt.Errorf("cannot resolve base path: %w", err)
	}
	joined := filepath.Join(absBase, child)
	cleaned := filepath.Clean(joined)

	// The cleaned path must be a direct child of absBase.
	if cleaned == absBase || !strings.HasPrefix(cleaned, absBase+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes base directory %q", child, base)
	}
	return cleaned, nil
}

// NoControlChars rejects any string containing ASCII bytes below 0x20 (space).
// This catches null bytes, newlines, carriage returns, tabs, etc.
func NoControlChars(s string) error {
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 {
			return fmt.Errorf("contains control character at byte %d (0x%02x)", i, s[i])
		}
	}
	return nil
}

var validWorkspaceName = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,128}$`)

// WorkspaceName validates that name contains only alphanumeric, dash, and underscore.
func WorkspaceName(name string) error {
	if !validWorkspaceName.MatchString(name) {
		return fmt.Errorf("must be 1-128 characters of [a-zA-Z0-9_-]")
	}
	return nil
}

// RepoSegment validates a single segment of a repository path (host, org, or repo name).
// Rejects empty strings, ".." path traversal, control characters, and dangerous characters.
func RepoSegment(s string) error {
	if s == "" {
		return fmt.Errorf("must not be empty")
	}
	if s == ".." || strings.HasPrefix(s, "../") || strings.HasSuffix(s, "/..") || strings.Contains(s, "/../") {
		return fmt.Errorf("must not contain path traversal")
	}
	if err := NoControlChars(s); err != nil {
		return err
	}
	if strings.ContainsAny(s, "?#%") {
		return fmt.Errorf("must not contain '?', '#', or '%%'")
	}
	return nil
}
