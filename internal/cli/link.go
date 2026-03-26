package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/Tiryoh/rgw/internal/alias"
	"github.com/Tiryoh/rgw/internal/config"
	"github.com/Tiryoh/rgw/internal/ghq"
	"github.com/Tiryoh/rgw/internal/selector"
	"github.com/Tiryoh/rgw/internal/symlink"
	"github.com/Tiryoh/rgw/internal/worktree"
)

func newLinkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Manage symlinks between worktrees and ROS workspaces",
	}
	cmd.AddCommand(newLinkSetCmd())
	cmd.AddCommand(newLinkStatusCmd())
	cmd.AddCommand(newLinkUnsetCmd())
	cmd.AddCommand(newLinkRepairCmd())
	return cmd
}

func newLinkSetCmd() *cobra.Command {
	var (
		pathFlag        string
		branchFlag      string
		interactiveFlag bool
	)
	cmd := &cobra.Command{
		Use:               "set <repo>",
		Short:             "Create or update a symlink for a repository",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeRepoArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			wsDef, err := cfg.ResolveWorkspace(flagWS)
			if err != nil {
				return err
			}
			ghqRoot, err := cfg.ResolveGHQRoot()
			if err != nil {
				return err
			}

			info, repoPath, err := ghq.ParseRepoArg(ghqRoot, args[0])
			if err != nil {
				return err
			}

			aliasMode, err := alias.Parse(cfg.ResolveAliasMode())
			if err != nil {
				return err
			}

			var targetPath string

			switch {
			case pathFlag != "":
				targetPath = pathFlag
			case branchFlag != "":
				wts, err := worktree.ListForRepo(repoPath)
				if err != nil {
					return err
				}
				var matched *worktree.Worktree
				for i := range wts {
					if wts[i].Branch == branchFlag {
						matched = &wts[i]
						break
					}
				}
				if matched == nil {
					fmt.Fprintf(os.Stderr, "Branch %q not found. Available branches:\n", branchFlag)
					for _, wt := range wts {
						fmt.Fprintf(os.Stderr, "  %s (%s)\n", wt.Branch, wt.Path)
					}
					return fmt.Errorf("branch %q not found in worktrees", branchFlag)
				}
				targetPath = matched.Path
			case interactiveFlag:
				wts, err := worktree.ListForRepo(repoPath)
				if err != nil {
					return err
				}
				selected, err := selector.Select(wts)
				if err != nil {
					return err
				}
				targetPath = selected.Path
			default:
				// Default: use the main worktree (first entry)
				wts, err := worktree.ListForRepo(repoPath)
				if err != nil {
					return err
				}
				if len(wts) == 0 {
					return fmt.Errorf("no worktrees found for %s", args[0])
				}
				targetPath = wts[0].Path
			}

			aliasName := alias.Resolve(info.Host, info.Org, info.Repo, aliasMode)

			if err := symlink.Set(wsDef, aliasName, targetPath); err != nil {
				return err
			}

			fmt.Printf("Linked %s -> %s\n", aliasName, targetPath)
			return nil
		},
	}
	cmd.Flags().StringVar(&pathFlag, "path", "", "worktree path to link directly")
	cmd.Flags().StringVar(&branchFlag, "branch", "", "select worktree by branch name")
	cmd.Flags().BoolVarP(&interactiveFlag, "interactive", "i", false, "interactively select worktree")
	return cmd
}

func newLinkStatusCmd() *cobra.Command {
	var allWS bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current link status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			var workspaces []config.WorkspaceDef
			if allWS {
				workspaces = cfg.ROS.Workspaces
			} else {
				wsDef, err := cfg.ResolveWorkspace(flagWS)
				if err != nil {
					return err
				}
				workspaces = []config.WorkspaceDef{*wsDef}
			}

			for _, ws := range workspaces {
				if allWS {
					fmt.Printf("\n=== %s (%s) ===\n", ws.Name, ws.Path)
				}
				links, err := symlink.Status(&ws)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
					continue
				}
				if len(links) == 0 {
					fmt.Println("No symlinks found.")
					continue
				}

				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "ALIAS\tTARGET\tSTATUS")
				for _, link := range links {
					status := "ok"
					if link.Orphaned {
						status = "BROKEN"
					}
					fmt.Fprintf(w, "%s\t%s\t%s\n", link.Alias, link.TargetPath, status)
				}
				w.Flush()
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&allWS, "all-ws", false, "show status for all workspaces")
	return cmd
}

func newLinkUnsetCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "unset <repo>",
		Short:             "Remove a symlink for a repository",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeRepoArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			wsDef, err := cfg.ResolveWorkspace(flagWS)
			if err != nil {
				return err
			}
			ghqRoot, err := cfg.ResolveGHQRoot()
			if err != nil {
				return err
			}

			info, _, err := ghq.ParseRepoArg(ghqRoot, args[0])
			if err != nil {
				return err
			}

			aliasMode, err := alias.Parse(cfg.ResolveAliasMode())
			if err != nil {
				return err
			}
			aliasName := alias.Resolve(info.Host, info.Org, info.Repo, aliasMode)

			if err := symlink.Unset(wsDef, aliasName); err != nil {
				return err
			}
			fmt.Printf("Unlinked %s\n", aliasName)
			return nil
		},
	}
}

func newLinkRepairCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "repair",
		Short: "Remove broken symlinks",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			wsDef, err := cfg.ResolveWorkspace(flagWS)
			if err != nil {
				return err
			}

			removed, err := symlink.Repair(wsDef)
			if err != nil {
				return err
			}
			if removed == 0 {
				fmt.Println("No broken symlinks found.")
			} else {
				fmt.Printf("Removed %d broken symlink(s).\n", removed)
			}
			return nil
		},
	}
}
