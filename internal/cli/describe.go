package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommandSchema is the JSON output of `rgw describe`.
type CommandSchema struct {
	Name        string          `json:"name"`
	FullPath    string          `json:"full_path"`
	Description string          `json:"description"`
	Arguments   []ArgSchema     `json:"arguments"`
	Flags       []FlagSchema    `json:"flags"`
	Subcommands []SubcmdSummary `json:"subcommands,omitempty"`
}

// ArgSchema describes a positional argument.
type ArgSchema struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
}

// FlagSchema describes a command flag.
type FlagSchema struct {
	Name        string `json:"name"`
	Shorthand   string `json:"shorthand,omitempty"`
	Type        string `json:"type"`
	Default     string `json:"default"`
	Description string `json:"description"`
	Persistent  bool   `json:"persistent"`
}

// SubcmdSummary is a brief description of a subcommand.
type SubcmdSummary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// argPattern matches both required <arg> and optional [arg] positional arguments.
var argPattern = regexp.MustCompile(`([<\[])(\w+)[>\]]`)

func newDescribeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "describe [command...]",
		Short: "Output machine-readable command metadata as JSON",
		Long:  "Walks the command tree and outputs structured JSON describing the target command's arguments, flags, and subcommands.",
		RunE: func(cmd *cobra.Command, args []string) error {
			target := cmd.Root()
			for _, name := range args {
				found := false
				for _, sub := range target.Commands() {
					if sub.Name() == name {
						target = sub
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("unknown command: %s", strings.Join(args, " "))
				}
			}

			schema := buildSchema(target)

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(schema)
		},
	}
}

func buildSchema(cmd *cobra.Command) CommandSchema {
	desc := cmd.Long
	if desc == "" {
		desc = cmd.Short
	}

	schema := CommandSchema{
		Name:        cmd.Name(),
		FullPath:    cmd.CommandPath(),
		Description: desc,
		Arguments:   extractArgs(cmd),
		Flags:       extractFlags(cmd),
	}

	for _, sub := range cmd.Commands() {
		if sub.IsAvailableCommand() {
			schema.Subcommands = append(schema.Subcommands, SubcmdSummary{
				Name:        sub.Name(),
				Description: sub.Short,
			})
		}
	}

	return schema
}

func extractArgs(cmd *cobra.Command) []ArgSchema {
	// Parse <arg> (required) and [arg] (optional) tokens from the Use string.
	matches := argPattern.FindAllStringSubmatch(cmd.Use, -1)
	args := make([]ArgSchema, 0, len(matches))
	for _, m := range matches {
		args = append(args, ArgSchema{
			Name:     m[2],
			Required: m[1] == "<",
		})
	}
	return args
}

func extractFlags(cmd *cobra.Command) []FlagSchema {
	var flags []FlagSchema

	// Local flags
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		flags = append(flags, FlagSchema{
			Name:        f.Name,
			Shorthand:   f.Shorthand,
			Type:        f.Value.Type(),
			Default:     f.DefValue,
			Description: f.Usage,
			Persistent:  false,
		})
	})

	// Inherited (persistent) flags
	cmd.InheritedFlags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		flags = append(flags, FlagSchema{
			Name:        f.Name,
			Shorthand:   f.Shorthand,
			Type:        f.Value.Type(),
			Default:     f.DefValue,
			Description: f.Usage,
			Persistent:  true,
		})
	})

	return flags
}
