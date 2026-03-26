package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Tiryoh/rgw/internal/config"
)

// completeRepoArg provides completion for <repo> arguments by listing
// repositories under ghq root in org/repo format.
func completeRepoArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	ghqRoot, err := cfg.ResolveGHQRoot()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	hosts, err := os.ReadDir(ghqRoot)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
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
			orgPath := filepath.Join(hostPath, org.Name())
			repos, err := os.ReadDir(orgPath)
			if err != nil {
				continue
			}
			for _, repo := range repos {
				if !repo.IsDir() || strings.HasPrefix(repo.Name(), ".") {
					continue
				}
				// Provide both short (org/repo) and full (host/org/repo) forms
				orgRepo := org.Name() + "/" + repo.Name()
				if strings.HasPrefix(orgRepo, toComplete) {
					completions = append(completions, orgRepo)
				}
				full := host.Name() + "/" + org.Name() + "/" + repo.Name()
				if strings.HasPrefix(full, toComplete) && !strings.HasPrefix(orgRepo, toComplete) {
					completions = append(completions, full)
				}
			}
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completeWSName provides completion for workspace name arguments.
func completeWSName(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, ws := range cfg.ROS.Workspaces {
		if strings.HasPrefix(ws.Name, toComplete) {
			completions = append(completions, ws.Name)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}
