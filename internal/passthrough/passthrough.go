package passthrough

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/jmelahman/sleight-of-hand/internal/lookup"
)

// Exec runs the real binary for toolName with args, forwarding all I/O
// and signals. Returns the exit code of the subprocess.
func Exec(toolName string, args []string) int {
	realPath, err := lookup.FindReal(toolName)
	if err != nil {
		_, _ = os.Stderr.WriteString("sleight-of-hand: cannot find real " + toolName + ": " + err.Error() + "\n")
		return 127
	}

	cmd := exec.Command(realPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sigCh {
			if cmd.Process != nil {
				_ = cmd.Process.Signal(sig)
			}
		}
	}()
	defer func() {
		signal.Stop(sigCh)
		close(sigCh)
	}()

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				if status.Signaled() {
					return 128 + int(status.Signal())
				}
				return status.ExitStatus()
			}
		}
		return 1
	}
	return 0
}
