package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/Tiryoh/rgw/internal/config"
	"github.com/Tiryoh/rgw/internal/ghq"
	"github.com/Tiryoh/rgw/internal/worktree"
)

func newWTCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wt",
		Short: "Worktree operations",
	}
	cmd.AddCommand(newWTListCmd())
	return cmd
}

func newWTListCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "list <repo>",
		Short:             "List worktrees for a repository",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeRepoArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			ghqRoot, err := cfg.ResolveGHQRoot()
			if err != nil {
				return err
			}
			_, repoPath, err := ghq.ParseRepoArg(ghqRoot, args[0])
			if err != nil {
				return err
			}

			wts, err := worktree.ListForRepo(repoPath)
			if err != nil {
				return err
			}
			if len(wts) == 0 {
				fmt.Println("No worktrees found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "BRANCH\tHEAD\tPATH")
			for _, wt := range wts {
				branch := wt.Branch
				if wt.Detached {
					branch = "(detached)"
				}
				head := wt.HEAD
				if len(head) > 8 {
					head = head[:8]
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", branch, head, wt.Path)
			}
			return w.Flush()
		},
	}
}
