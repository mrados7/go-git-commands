package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func CheckoutNewBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func CheckIfGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err == nil
}

func GetStagedFiles() []string {
	cmd := exec.Command("git", "diff", "--name-only", "--cached")
	out, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	if len(out) == 0 {
		return []string{}
	}

	return filterEmpty(strings.Split(string(out), "\n"))
}

func filterEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func GetCurrentGitBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error getting branch name: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}
