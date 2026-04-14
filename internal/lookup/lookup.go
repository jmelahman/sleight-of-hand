package lookup

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindReal locates the real binary for toolName by walking PATH,
// skipping any entry that resolves to the same executable as the
// current process (i.e., the sleight-of-hand binary itself).
func FindReal(toolName string) (string, error) {
	selfPath, err := selfResolved()
	if err != nil {
		return "", fmt.Errorf("cannot resolve self: %w", err)
	}

	pathEnv := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(pathEnv) {
		candidate := filepath.Join(dir, toolName)
		resolved, err := filepath.EvalSymlinks(candidate)
		if err != nil {
			continue
		}
		if resolved == selfPath {
			continue
		}
		info, err := os.Stat(resolved)
		if err != nil || info.IsDir() {
			continue
		}
		if info.Mode()&0111 != 0 {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("%q not found in PATH (excluding self)", toolName)
}

func selfResolved() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}
