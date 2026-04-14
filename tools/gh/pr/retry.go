package pr

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func newRetryCmd() *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "retry <pr-number>",
		Short: "Retry all failed jobs for a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRetry(args[0], repo)
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "R", "", "repository in OWNER/REPO format")

	return cmd
}

type prInfo struct {
	HeadRefName string `json:"headRefName"`
}

type runInfo struct {
	DatabaseID int    `json:"databaseId"`
	Conclusion string `json:"conclusion"`
	Name       string `json:"name"`
}

func runRetry(prNumber, repo string) error {
	branch, err := getPRBranch(prNumber, repo)
	if err != nil {
		return fmt.Errorf("getting PR branch: %w", err)
	}

	runs, err := getFailedRuns(branch, repo)
	if err != nil {
		return fmt.Errorf("listing runs: %w", err)
	}

	if len(runs) == 0 {
		fmt.Println("No failed runs found.")
		return nil
	}

	var errs []string
	for _, run := range runs {
		fmt.Printf("Retrying failed jobs in run %q (ID: %d)...\n", run.Name, run.DatabaseID)
		if err := rerunFailed(run.DatabaseID, repo); err != nil {
			errs = append(errs, fmt.Sprintf("  run %d: %v", run.DatabaseID, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("some reruns failed:\n%s", strings.Join(errs, "\n"))
	}
	return nil
}

func getPRBranch(prNumber, repo string) (string, error) {
	args := []string{"pr", "view", prNumber, "--json", "headRefName"}
	if repo != "" {
		args = append(args, "--repo", repo)
	}

	out, err := ghExec(args...)
	if err != nil {
		return "", err
	}

	var info prInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return "", fmt.Errorf("parsing PR info: %w", err)
	}
	return info.HeadRefName, nil
}

func getFailedRuns(branch, repo string) ([]runInfo, error) {
	args := []string{"run", "list", "--branch", branch, "--json", "databaseId,conclusion,name"}
	if repo != "" {
		args = append(args, "--repo", repo)
	}

	out, err := ghExec(args...)
	if err != nil {
		return nil, err
	}

	var runs []runInfo
	if err := json.Unmarshal(out, &runs); err != nil {
		return nil, fmt.Errorf("parsing run list: %w", err)
	}

	var failed []runInfo
	for _, r := range runs {
		if r.Conclusion == "failure" {
			failed = append(failed, r)
		}
	}
	return failed, nil
}

func rerunFailed(runID int, repo string) error {
	args := []string{"run", "rerun", fmt.Sprint(runID), "--failed"}
	if repo != "" {
		args = append(args, "--repo", repo)
	}

	cmd := exec.Command("gh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ghExec runs the real gh CLI and captures stdout.
func ghExec(args ...string) ([]byte, error) {
	cmd := exec.Command("gh", args...)
	cmd.Stderr = os.Stderr
	return cmd.Output()
}
