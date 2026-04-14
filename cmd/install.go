package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jmelahman/sleight-of-hand/internal/dispatch"
	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	var binDir string

	cmd := &cobra.Command{
		Use:   "install [tools...]",
		Short: "Create entrypoint symlinks for wrapped tools",
		Long: `Creates a symlink for each wrapped tool in the target directory.
When no tools are specified, all registered tools are installed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall(binDir, args)
		},
	}

	defaultBinDir := filepath.Join(os.Getenv("HOME"), ".local", "bin")
	cmd.Flags().StringVar(&binDir, "bin-dir", defaultBinDir, "directory in which to create symlinks")

	return cmd
}

func runInstall(binDir string, tools []string) error {
	selfPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot resolve self: %w", err)
	}
	selfPath, err = filepath.EvalSymlinks(selfPath)
	if err != nil {
		return fmt.Errorf("cannot resolve self symlink: %w", err)
	}

	if len(tools) == 0 {
		tools = dispatch.RegisteredTools()
	}

	for _, tool := range tools {
		if !dispatch.IsRegistered(tool) {
			return fmt.Errorf("unknown tool %q; registered tools: %s", tool, strings.Join(dispatch.RegisteredTools(), ", "))
		}
	}

	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("cannot create bin dir %s: %w", binDir, err)
	}

	// Warn if binDir is not in PATH.
	if !slices.Contains(filepath.SplitList(os.Getenv("PATH")), binDir) {
		fmt.Fprintf(os.Stderr, "warning: %s is not in PATH; add it with:\n  export PATH=%q:$PATH\n", binDir, binDir)
	}

	for _, tool := range tools {
		linkPath := filepath.Join(binDir, tool)

		if info, err := os.Lstat(linkPath); err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				if err := os.Remove(linkPath); err != nil {
					return fmt.Errorf("cannot remove existing symlink %s: %w", linkPath, err)
				}
			} else {
				return fmt.Errorf("%s already exists and is not a symlink; refusing to overwrite", linkPath)
			}
		}

		if err := os.Symlink(selfPath, linkPath); err != nil {
			return fmt.Errorf("cannot create symlink %s -> %s: %w", linkPath, selfPath, err)
		}
		fmt.Printf("installed: %s -> %s\n", linkPath, selfPath)
	}

	return nil
}
