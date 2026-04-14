package gh

import (
	"os"

	"github.com/jmelahman/sleight-of-hand/internal/passthrough"
	"github.com/jmelahman/sleight-of-hand/tools/gh/pr"
	"github.com/spf13/cobra"
)

// Run is the entry point called by dispatch when os.Args[0] is "gh".
func Run(args []string) int {
	rootCmd := newGhCmd()
	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		return 1
	}
	return 0
}

func newGhCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "gh",
		Short:              "GitHub CLI wrapper with extended commands",
		SilenceErrors:      true,
		SilenceUsage:       true,
		DisableFlagParsing: true,
		RunE: func(c *cobra.Command, args []string) error {
			os.Exit(passthrough.Exec("gh", args))
			return nil
		},
	}

	cmd.AddCommand(pr.NewPrCmd())

	return cmd
}
