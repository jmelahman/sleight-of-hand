package pr

import (
	"os"

	"github.com/jmelahman/sleight-of-hand/internal/passthrough"
	"github.com/spf13/cobra"
)

// NewPrCmd returns the "pr" subcommand group with custom extensions
// and a catch-all passthrough for unrecognized subcommands.
func NewPrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "pr",
		Short:              "Pull request commands (with extensions)",
		SilenceErrors:      true,
		SilenceUsage:       true,
		DisableFlagParsing: true,
		RunE: func(c *cobra.Command, args []string) error {
			os.Exit(passthrough.Exec("gh", append([]string{"pr"}, args...)))
			return nil
		},
	}

	cmd.AddCommand(newRetryCmd())

	return cmd
}
