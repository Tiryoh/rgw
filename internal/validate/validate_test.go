package validate

import (
	"testing"
)

func TestSafePath(t *testing.T) {
	base := t.TempDir()

	tests := []struct {
		name    string
		child   string
		wantErr bool
	}{
		{"simple child", "foo", false},
		{"nested child", "foo/bar", false},
		{"double underscore alias", "Tiryoh__my_pkg", false},
		{"traversal up", "../etc/passwd", true},
		{"traversal nested", "foo/../../etc", true},
		{"current dir only", ".", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SafePath(base, tt.child)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafePath(%q, %q) error = %v, wantErr = %v", base, tt.child, err, tt.wantErr)
			}
		})
	}
}

func TestNoControlChars(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"normal text", "hello world", false},
		{"with dash underscore dot", "a-b_c.d", false},
		{"empty string", "", false},
		{"with null byte", "foo\x00bar", true},
		{"with newline", "foo\nbar", true},
		{"with carriage return", "foo\rbar", true},
		{"with tab", "foo\tbar", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NoControlChars(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NoControlChars(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestWorkspaceName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"simple", "ros_ws", false},
		{"with dash", "my-workspace", false},
		{"alphanumeric", "ws1", false},
		{"empty", "", true},
		{"with space", "foo bar", true},
		{"with slash", "foo/bar", true},
		{"with dot dot", "a..b", true},
		{"too long", string(make([]byte, 129)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WorkspaceName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("WorkspaceName(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestRepoSegment(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"hostname", "github.com", false},
		{"org name", "Tiryoh", false},
		{"repo with underscore", "my_robot", false},
		{"empty", "", true},
		{"dot dot", "..", true},
		{"question mark", "foo?bar", true},
		{"hash", "foo#bar", true},
		{"percent", "foo%bar", true},
		{"null byte", "foo\x00bar", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RepoSegment(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("RepoSegment(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
