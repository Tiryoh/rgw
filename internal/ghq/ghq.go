package ghq

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RepoInfo represents a repository identified by its ghq-style path components.
type RepoInfo struct {
	Host string // e.g. "github.com"
	Org  string // e.g. "Tiryoh"
	Repo string // e.g. "rgw"
}

// ParseRepoArg takes the <repo> CLI argument and resolves it to a RepoInfo
// and the absolute filesystem path. Supported formats:
//   - "repo" — searches ghq root for matching repo name
//   - "org/repo" — assumes github.com
//   - "host/org/repo" — fully qualified
func ParseRepoArg(ghqRoot string, repoArg string) (*RepoInfo, string, error) {
	parts := strings.Split(repoArg, "/")
	switch len(parts) {
	case 1:
		// Search for repo name
		matches, err := FindRepoPaths(ghqRoot, repoArg)
		if err != nil {
			return nil, "", err
		}
		if len(matches) == 0 {
			return nil, "", fmt.Errorf("repository %q not found under %s", repoArg, ghqRoot)
		}
		if len(matches) > 1 {
			return nil, "", fmt.Errorf("ambiguous repository %q, found %d matches:\n  %s\nSpecify as org/repo or host/org/repo",
				repoArg, len(matches), strings.Join(matches, "\n  "))
		}
		info := infoFromPath(ghqRoot, matches[0])
		return info, matches[0], nil

	case 2:
		// org/repo -> assume github.com
		info := &RepoInfo{Host: "github.com", Org: parts[0], Repo: parts[1]}
		p := filepath.Join(ghqRoot, info.Host, info.Org, info.Repo)
		if _, err := os.Stat(p); err != nil {
			return nil, "", fmt.Errorf("repository not found at %s", p)
		}
		return info, p, nil

	case 3:
		// host/org/repo
		info := &RepoInfo{Host: parts[0], Org: parts[1], Repo: parts[2]}
		p := filepath.Join(ghqRoot, info.Host, info.Org, info.Repo)
		if _, err := os.Stat(p); err != nil {
			return nil, "", fmt.Errorf("repository not found at %s", p)
		}
		return info, p, nil

	default:
		return nil, "", fmt.Errorf("invalid repository argument %q", repoArg)
	}
}

// FindRepoPaths searches the ghq root for directories whose name matches repoName.
func FindRepoPaths(ghqRoot string, repoName string) ([]string, error) {
	var matches []string
	hosts, err := os.ReadDir(ghqRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to read ghq root %s: %w", ghqRoot, err)
	}
	for _, host := range hosts {
		if !host.IsDir() || strings.HasPrefix(host.Name(), ".") {
			continue
		}
		hostPath := filepath.Join(ghqRoot, host.Name())
		orgs, err := os.ReadDir(hostPath)
		if err != nil {
			continue
		}
		for _, org := range orgs {
			if !org.IsDir() || strings.HasPrefix(org.Name(), ".") {
				continue
			}
			repoPath := filepath.Join(hostPath, org.Name(), repoName)
			if info, err := os.Stat(repoPath); err == nil && info.IsDir() {
				matches = append(matches, repoPath)
			}
		}
	}
	return matches, nil
}

func infoFromPath(ghqRoot, repoPath string) *RepoInfo {
	rel, _ := filepath.Rel(ghqRoot, repoPath)
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) >= 3 {
		return &RepoInfo{Host: parts[0], Org: parts[1], Repo: parts[2]}
	}
	return &RepoInfo{Repo: filepath.Base(repoPath)}
}
