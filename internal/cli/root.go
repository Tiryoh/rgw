package cli

import (
	"github.com/spf13/cobra"
)

var (
	flagWS      string
	flagVerbose bool
	flagDryRun  bool
)

// NewRootCmd creates the root "rgw" command with all subcommands.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "rgw",
		Short: "ROS workspace view controller for git worktrees",
		Long:  "rgw detects git worktrees and manages symlinks in ROS workspace src/ directories.",
	}

	rootCmd.PersistentFlags().StringVarP(&flagWS, "ws", "w", "", "override active workspace")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "", `output format: "text" or "json" (default: auto-detect by TTY)`)
	rootCmd.PersistentFlags().BoolVar(&flagDryRun, "dry-run", false, "show what would be done without making changes")

	rootCmd.AddCommand(newWSCmd())
	rootCmd.AddCommand(newWTCmd())
	rootCmd.AddCommand(newLinkCmd())
	rootCmd.AddCommand(newOpenCmd())
	rootCmd.AddCommand(newDoctorCmd())
	rootCmd.AddCommand(newDescribeCmd())

	return rootCmd
}
