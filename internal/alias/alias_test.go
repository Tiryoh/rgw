package alias

import (
	"testing"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		host, org, repo string
		mode            Mode
		want            string
	}{
		{"github.com", "Tiryoh", "my_pkg", ModeRepo, "my_pkg"},
		{"github.com", "Tiryoh", "my_pkg", ModeOrgRepo, "Tiryoh__my_pkg"},
		{"github.com", "Tiryoh", "my_pkg", ModeHostOrgRepo, "github.com__Tiryoh__my_pkg"},
		{"gitlab.com", "org", "repo", ModeOrgRepo, "org__repo"},
	}
	for _, tt := range tests {
		got, err := Resolve(tt.host, tt.org, tt.repo, tt.mode)
		if err != nil {
			t.Errorf("Resolve(%q,%q,%q,%q) unexpected error: %v",
				tt.host, tt.org, tt.repo, tt.mode, err)
			continue
		}
		if got != tt.want {
			t.Errorf("Resolve(%q,%q,%q,%q) = %q, want %q",
				tt.host, tt.org, tt.repo, tt.mode, got, tt.want)
		}
	}
}

func TestParse(t *testing.T) {
	for _, valid := range []string{"repo", "org_repo", "host_org_repo"} {
		if _, err := Parse(valid); err != nil {
			t.Errorf("Parse(%q) error: %v", valid, err)
		}
	}
	if _, err := Parse("invalid"); err == nil {
		t.Error("expected error for invalid mode")
	}
}
