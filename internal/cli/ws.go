package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/Tiryoh/rgw/internal/config"
	"github.com/Tiryoh/rgw/internal/workspace"
)

func newWSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ws",
		Short: "Manage ROS workspaces",
	}
	cmd.AddCommand(newWSListCmd())
	cmd.AddCommand(newWSAddCmd())
	cmd.AddCommand(newWSUseCmd())
	cmd.AddCommand(newWSCurrentCmd())
	return cmd
}

func newWSListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured workspaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			wsList := workspace.List(cfg)
			current, _ := cfg.ResolveWorkspace(flagWS)

			if isJSON() {
				type wsEntry struct {
					Name   string `json:"name"`
					Path   string `json:"path"`
					Active bool   `json:"active"`
				}
				entries := make([]wsEntry, 0, len(wsList))
				for _, ws := range wsList {
					entries = append(entries, wsEntry{
						Name:   ws.Name,
						Path:   ws.Path,
						Active: current != nil && ws.Name == current.Name,
					})
				}
				return printJSON(entries)
			}

			if len(wsList) == 0 {
				fmt.Println("No workspaces configured. Use 'rgw ws add' to add one.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tPATH\tACTIVE")
			for _, ws := range wsList {
				active := ""
				if current != nil && ws.Name == current.Name {
					active = "*"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", ws.Name, ws.Path, active)
			}
			return w.Flush()
		},
	}
}

func newWSAddCmd() *cobra.Command {
	var name, path string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if flagDryRun {
				return printAction(fmt.Sprintf("[dry-run] Would add workspace %q at %s", name, path))
			}
			if err := workspace.Add(cfg, name, path); err != nil {
				return err
			}
			return printAction(fmt.Sprintf("Added workspace %q at %s", name, path))
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "workspace name (required)")
	cmd.Flags().StringVar(&path, "path", "", "workspace path (required)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("path")
	return cmd
}

func newWSUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "use <name>",
		Short:             "Set the default workspace",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeWSName,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if flagDryRun {
				return printAction(fmt.Sprintf("[dry-run] Would set default workspace to %q", args[0]))
			}
			if err := workspace.Use(cfg, args[0]); err != nil {
				return err
			}
			return printAction(fmt.Sprintf("Default workspace set to %q", args[0]))
		},
	}
}

func newWSCurrentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show the current active workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			ws, err := workspace.Current(cfg, flagWS)
			if err != nil {
				return err
			}
			if isJSON() {
				return printJSON(ws)
			}
			fmt.Printf("%s (%s)\n", ws.Name, ws.Path)
			return nil
		},
	}
}
