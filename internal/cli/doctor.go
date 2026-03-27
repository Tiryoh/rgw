package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Tiryoh/rgw/internal/config"
	"github.com/Tiryoh/rgw/internal/doctor"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check environment health",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			results := doctor.RunAll(cfg)

			if isJSON() {
				return printJSON(results)
			}

			hasErrors := false
			for _, r := range results {
				prefix := fmt.Sprintf("[%-4s]", r.Severity)
				fmt.Printf("%s  %s: %s\n", prefix, r.Name, r.Message)
				if r.Severity == doctor.SeverityError {
					hasErrors = true
				}
			}

			if hasErrors {
				fmt.Println("\nSome checks failed. Run 'rgw link repair' to fix broken symlinks.")
			} else {
				fmt.Println("\nAll checks passed.")
			}
			return nil
		},
	}
}
