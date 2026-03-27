package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
)

var flagOutput string

// outputFormat returns the resolved output format ("text" or "json").
// If the user explicitly set --output, that wins. Otherwise, default
// to "json" when stdout is not a TTY.
func outputFormat() string {
	if flagOutput != "" {
		return flagOutput
	}
	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return "text"
	}
	return "json"
}

func isJSON() bool {
	return outputFormat() == "json"
}

// printJSON encodes v as indented JSON to stdout.
func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

type actionResult struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// printAction prints a status message in the current format.
func printAction(msg string) error {
	if isJSON() {
		return printJSON(actionResult{OK: true, Message: msg})
	}
	fmt.Println(msg)
	return nil
}
