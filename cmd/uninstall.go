package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmelahman/sleight-of-hand/internal/dispatch"
	"github.com/spf13/cobra"
)

func newUninstallCmd() *cobra.Command {
	var binDir string

	cmd := &cobra.Command{
		Use:   "uninstall [tools...]",
		Short: "Remove entrypoint symlinks for wrapped tools",
		Long: `Removes symlinks for each wrapped tool from the target directory.
Only removes symlinks that point back to the sleight-of-hand binary.
When no tools are specified, all registered tools are uninstalled.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUninstall(binDir, args)
		},
	}

	defaultBinDir := filepath.Join(os.Getenv("HOME"), ".local", "bin")
	cmd.Flags().StringVar(&binDir, "bin-dir", defaultBinDir, "directory from which to remove symlinks")

	return cmd
}

func runUninstall(binDir string, tools []string) error {
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
		linkPath := filepath.Join(binDir, tool)

		info, err := os.Lstat(linkPath)
		if err != nil {
			fmt.Printf("skipped: %s (not found)\n", linkPath)
			continue
		}

		if info.Mode()&os.ModeSymlink == 0 {
			fmt.Printf("skipped: %s (not a symlink)\n", linkPath)
			continue
		}

		target, err := os.Readlink(linkPath)
		if err != nil {
			return fmt.Errorf("cannot read symlink %s: %w", linkPath, err)
		}

		resolvedTarget, err := filepath.EvalSymlinks(target)
		if err != nil {
			resolvedTarget = target
		}

		if resolvedTarget != selfPath {
			fmt.Printf("skipped: %s (points to %s, not sleight-of-hand)\n", linkPath, target)
			continue
		}

		if err := os.Remove(linkPath); err != nil {
			return fmt.Errorf("cannot remove %s: %w", linkPath, err)
		}
		fmt.Printf("uninstalled: %s\n", linkPath)
	}

	return nil
}
