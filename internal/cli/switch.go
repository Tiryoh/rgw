package cli

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/Tiryoh/rgw/internal/config"
	"github.com/Tiryoh/rgw/internal/selector"
	"github.com/Tiryoh/rgw/internal/symlink"
	"github.com/Tiryoh/rgw/internal/validate"
	"github.com/Tiryoh/rgw/internal/worktree"
)

func newSwitchCmd() *cobra.Command {
	var (
		pathFlag        string
		interactiveFlag bool
	)
	cmd := &cobra.Command{
		Use:   "switch <alias> [branch]",
		Short: "Switch an existing symlink to a different worktree",
		Long:  "Switch the target of an existing symlink to a different worktree branch. Takes the alias name as shown in 'rgw link status'.",
		Args:  cobra.RangeArgs(1, 2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			switch len(args) {
			case 0:
				return completeAliasArg(cmd, args, toComplete)
			case 1:
				return completeSwitchBranchArg(cmd, args, toComplete)
			default:
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]

			// Validate alias input at boundary.
			if err := validate.RepoSegment(alias); err != nil {
				return fmt.Errorf("invalid alias %q: %w", alias, err)
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}
			wsDef, err := cfg.ResolveWorkspace(flagWS)
			if err != nil {
				return err
			}

			link, err := symlink.FindByAlias(wsDef, alias)
			if err != nil {
				return err
			}

			// Determine branch from second positional arg or --branch flag (kept for compatibility).
			branch := ""
			if len(args) >= 2 {
				branch = args[1]
			}

			var targetPath string

			switch {
			case pathFlag != "":
				if err := validate.NoControlChars(pathFlag); err != nil {
					return fmt.Errorf("invalid --path %q: %w", pathFlag, err)
				}
				targetPath = pathFlag
			case branch != "":
				if link.Orphaned {
					return fmt.Errorf("current target %s does not exist; use --path to specify a new target directly", link.TargetPath)
				}
				wts, err := worktree.ListForRepo(link.TargetPath)
				if err != nil {
					return fmt.Errorf("failed to list worktrees: %w", err)
				}
				var matched *worktree.Worktree
				for i := range wts {
					if wts[i].Branch == branch {
						matched = &wts[i]
						break
					}
				}
				if matched == nil {
					fmt.Fprintf(os.Stderr, "Branch %q not found. Available branches:\n", branch)
					for _, wt := range wts {
						fmt.Fprintf(os.Stderr, "  %s (%s)\n", wt.Branch, wt.Path)
					}
					return fmt.Errorf("branch %q not found in worktrees", branch)
				}
				targetPath = matched.Path
			case interactiveFlag:
				if link.Orphaned {
					return fmt.Errorf("current target %s does not exist; use --path to specify a new target directly", link.TargetPath)
				}
				wts, err := worktree.ListForRepo(link.TargetPath)
				if err != nil {
					return fmt.Errorf("failed to list worktrees: %w", err)
				}
				selected, err := selector.Select(wts)
				if err != nil {
					return err
				}
				targetPath = selected.Path
			default:
				// No args/flags: require branch in non-TTY; default to interactive in TTY.
				if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
					return fmt.Errorf("no branch specified; in non-TTY mode, use: rgw switch <alias> <branch>")
				}
				if link.Orphaned {
					return fmt.Errorf("current target %s does not exist; use --path to specify a new target directly", link.TargetPath)
				}
				wts, err := worktree.ListForRepo(link.TargetPath)
				if err != nil {
					return fmt.Errorf("failed to list worktrees: %w", err)
				}
				selected, err := selector.Select(wts)
				if err != nil {
					return err
				}
				targetPath = selected.Path
			}

			if targetPath == link.TargetPath {
				return printAction(fmt.Sprintf("%s is already linked to %s", alias, targetPath))
			}

			if flagDryRun {
				if isJSON() {
					return printJSON(actionResult{OK: true, Message: fmt.Sprintf("[dry-run] Would switch %s: %s -> %s", alias, link.TargetPath, targetPath)})
				}
				fmt.Printf("[dry-run] Would switch %s:\n  %s\n  -> %s\n", alias, link.TargetPath, targetPath)
				return nil
			}

			if err := symlink.Set(wsDef, alias, targetPath); err != nil {
				return err
			}

			if isJSON() {
				return printJSON(actionResult{OK: true, Message: fmt.Sprintf("Switched %s: %s -> %s", alias, link.TargetPath, targetPath)})
			}
			fmt.Printf("Switched %s:\n  %s\n  -> %s\n", alias, link.TargetPath, targetPath)
			return nil
		},
	}
	cmd.Flags().StringVar(&pathFlag, "path", "", "target path to switch to directly")
	cmd.Flags().BoolVarP(&interactiveFlag, "interactive", "i", false, "interactively select worktree")
	return cmd
}
