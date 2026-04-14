package main

import (
	"os"
	"path/filepath"

	"github.com/jmelahman/sleight-of-hand/cmd"
	"github.com/jmelahman/sleight-of-hand/internal/dispatch"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	toolName := filepath.Base(os.Args[0])
	if toolName == "sleight-of-hand" {
		rootCmd := cmd.NewRootCmd(version, commit)
		cobra.CheckErr(rootCmd.Execute())
		return
	}
	os.Exit(dispatch.Run(toolName, os.Args[1:]))
}
