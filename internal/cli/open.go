package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/Tiryoh/rgw/internal/config"
	"github.com/Tiryoh/rgw/internal/ghq"
	"github.com/Tiryoh/rgw/internal/selector"
	"github.com/Tiryoh/rgw/internal/worktree"
)

func newOpenCmd() *cobra.Command {
	var (
		pathFlag        string
		interactiveFlag bool
	)
	cmd := &cobra.Command{
		Use:               "open <repo>",
		Short:             "Open a worktree in an editor",
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

			var targetPath string

			switch {
			case pathFlag != "":
				targetPath = pathFlag
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
				// Open the main repo path
				targetPath = repoPath
			}

			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "code"
			}

			fmt.Printf("Opening %s with %s\n", targetPath, editor)
			c := exec.Command(editor, targetPath)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}
	cmd.Flags().StringVar(&pathFlag, "path", "", "worktree path to open directly")
	cmd.Flags().BoolVarP(&interactiveFlag, "interactive", "i", false, "interactively select worktree")
	return cmd
}
