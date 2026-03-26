package alias

import "fmt"

// Mode represents an alias naming strategy.
type Mode string

const (
	ModeRepo        Mode = "repo"
	ModeOrgRepo     Mode = "org_repo"
	ModeHostOrgRepo Mode = "host_org_repo"
)

const separator = "__"

// Resolve computes the alias string for a given repository identity and mode.
//
// Examples for github.com/Tiryoh/my_pkg:
//
//	ModeRepo        -> "my_pkg"
//	ModeOrgRepo     -> "Tiryoh__my_pkg"
//	ModeHostOrgRepo -> "github.com__Tiryoh__my_pkg"
func Resolve(host, org, repo string, mode Mode) string {
	switch mode {
	case ModeRepo:
		return repo
	case ModeHostOrgRepo:
		return host + separator + org + separator + repo
	default: // ModeOrgRepo
		return org + separator + repo
	}
}

// Parse validates and returns a Mode from a string.
func Parse(s string) (Mode, error) {
	switch Mode(s) {
	case ModeRepo, ModeOrgRepo, ModeHostOrgRepo:
		return Mode(s), nil
	default:
		return "", fmt.Errorf("unknown alias mode %q (valid: repo, org_repo, host_org_repo)", s)
	}
}
