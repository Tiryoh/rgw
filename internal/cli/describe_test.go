package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestExtractArgsRequired(t *testing.T) {
	cmd := &cobra.Command{
		Use:  "set <repo>",
		Args: cobra.ExactArgs(1),
	}
	args := extractArgs(cmd)
	if len(args) != 1 {
		t.Fatalf("len(args) = %d, want 1", len(args))
	}
	if args[0].Name != "repo" {
		t.Errorf("args[0].Name = %q, want %q", args[0].Name, "repo")
	}
	if !args[0].Required {
		t.Error("args[0].Required = false, want true")
	}
}

func TestExtractArgsOptional(t *testing.T) {
	cmd := &cobra.Command{
		Use:  "switch <alias> [branch]",
		Args: cobra.RangeArgs(1, 2),
	}
	args := extractArgs(cmd)
	if len(args) != 2 {
		t.Fatalf("len(args) = %d, want 2", len(args))
	}
	if args[0].Name != "alias" || !args[0].Required {
		t.Errorf("args[0] = %+v, want {Name:alias Required:true}", args[0])
	}
	if args[1].Name != "branch" || args[1].Required {
		t.Errorf("args[1] = %+v, want {Name:branch Required:false}", args[1])
	}
}

func TestExtractArgsNoArgs(t *testing.T) {
	cmd := &cobra.Command{
		Use: "status",
	}
	args := extractArgs(cmd)
	if len(args) != 0 {
		t.Errorf("len(args) = %d, want 0", len(args))
	}
}
