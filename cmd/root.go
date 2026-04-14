package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewRootCmd(version, commit string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sleight-of-hand",
		Short:   "Transparent CLI wrapper with custom extensions",
		Version: fmt.Sprintf("%s\ncommit %s", version, commit),
	}

	cmd.AddCommand(newInstallCmd())
	cmd.AddCommand(newUninstallCmd())

	return cmd
}
