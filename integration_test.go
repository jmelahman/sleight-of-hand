package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompletionProxiesToRealBinary(t *testing.T) {
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		t.Skip("gh not found in PATH")
	}

	// Build the shim binary.
	tmpDir := t.TempDir()
	shimBin := filepath.Join(tmpDir, "sleight-of-hand")
	build := exec.Command("go", "build", "-o", shimBin, ".")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		t.Fatalf("building shim: %v", err)
	}

	// Create a symlink so the shim intercepts "gh".
	shimGh := filepath.Join(tmpDir, "gh")
	if err := os.Symlink(shimBin, shimGh); err != nil {
		t.Fatalf("creating symlink: %v", err)
	}

	// Build a PATH with our shim directory first, followed by the real gh's directory.
	realGhDir := filepath.Dir(ghPath)
	testPATH := tmpDir + string(os.PathListSeparator) + realGhDir

	// Build an environment with only our controlled PATH to avoid
	// stale symlinks or other PATH entries interfering.
	var testEnv []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "PATH=") {
			testEnv = append(testEnv, e)
		}
	}
	testEnv = append(testEnv, "PATH="+testPATH)

	runComplete := func(args ...string) string {
		t.Helper()
		cmdArgs := append([]string{"__complete"}, args...)
		cmd := exec.Command(shimGh, cmdArgs...)
		cmd.Env = testEnv
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("running %v: %v", cmdArgs, err)
		}
		return string(out)
	}

	completionNames := func(output string) []string {
		var names []string
		for _, line := range strings.Split(strings.TrimRight(output, "\n"), "\n") {
			if strings.HasPrefix(line, ":") {
				continue // directive line
			}
			name := strings.SplitN(line, "\t", 2)[0]
			names = append(names, name)
		}
		return names
	}

	hasName := func(names []string, want string) bool {
		for _, n := range names {
			if n == want {
				return true
			}
		}
		return false
	}

	t.Run("root completions include real gh subcommands", func(t *testing.T) {
		out := runComplete("")
		names := completionNames(out)

		// These are core gh subcommands that should always exist.
		for _, want := range []string{"issue", "repo", "auth"} {
			if !hasName(names, want) {
				t.Errorf("missing real gh subcommand %q in root completions:\n%s", want, out)
			}
		}
	})

	t.Run("root completions include shim subcommands", func(t *testing.T) {
		out := runComplete("")
		names := completionNames(out)

		if !hasName(names, "pr") {
			t.Errorf("missing shim subcommand %q in root completions:\n%s", "pr", out)
		}
	})

	t.Run("pr completions include both shim and real subcommands", func(t *testing.T) {
		out := runComplete("pr", "")
		names := completionNames(out)

		if !hasName(names, "retry") {
			t.Errorf("missing shim subcommand %q in pr completions:\n%s", "retry", out)
		}
		for _, want := range []string{"list", "view", "create", "merge"} {
			if !hasName(names, want) {
				t.Errorf("missing real gh pr subcommand %q in pr completions:\n%s", want, out)
			}
		}
	})

	t.Run("non-shimmed subcommand completions proxy through", func(t *testing.T) {
		out := runComplete("issue", "")
		names := completionNames(out)

		for _, want := range []string{"list", "view", "create"} {
			if !hasName(names, want) {
				t.Errorf("missing real gh issue subcommand %q in completions:\n%s", want, out)
			}
		}
	})

	t.Run("output ends with a valid directive line", func(t *testing.T) {
		out := runComplete("")
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		last := lines[len(lines)-1]
		if !strings.HasPrefix(last, ":") {
			t.Errorf("last line should be a directive (e.g. ':4'), got %q", last)
		}
	})
}
