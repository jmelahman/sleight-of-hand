package completion

import (
	"os/exec"
	"strings"

	"github.com/jmelahman/sleight-of-hand/internal/lookup"
	"github.com/spf13/cobra"
)

// ValidArgsFromReal returns a ValidArgsFunction that proxies completion requests
// to the real underlying binary found in PATH. The prefix args are prepended to
// the completion args (e.g., ["pr"] for "gh pr" subcommands).
func ValidArgsFromReal(toolName string, prefixArgs ...string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		realPath, err := lookup.FindReal(toolName)
		if err != nil {
			return nil, cobra.ShellCompDirectiveDefault
		}

		// Build: <real-binary> __complete [prefixArgs...] [args...] <toComplete>
		completeArgs := []string{"__complete"}
		completeArgs = append(completeArgs, prefixArgs...)
		completeArgs = append(completeArgs, args...)
		completeArgs = append(completeArgs, toComplete)

		out, err := exec.Command(realPath, completeArgs...).Output()
		if err != nil {
			return nil, cobra.ShellCompDirectiveDefault
		}

		return parseCompletionOutput(out, cmd)
	}
}

func parseCompletionOutput(out []byte, cmd *cobra.Command) ([]string, cobra.ShellCompDirective) {
	// Collect known subcommand names so we can skip duplicates.
	shimNames := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		if sub.IsAvailableCommand() {
			shimNames[sub.Name()] = true
		}
	}

	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	if len(lines) == 0 {
		return nil, cobra.ShellCompDirectiveDefault
	}

	// The last line is the directive (e.g. ":4").
	directiveLine := lines[len(lines)-1]
	lines = lines[:len(lines)-1]

	var completions []string
	for _, line := range lines {
		name := strings.SplitN(line, "\t", 2)[0]
		if shimNames[name] {
			continue
		}
		completions = append(completions, line)
	}

	directive := parseDirective(directiveLine)
	return completions, directive
}

func parseDirective(s string) cobra.ShellCompDirective {
	s = strings.TrimPrefix(s, ":")
	var d int
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			d = d*10 + int(ch-'0')
		}
	}
	return cobra.ShellCompDirective(d)
}
