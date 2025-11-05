// Package prompt provides utilities to build and render the interactive shell
// prompt. It handles displaying the current working directory (abbreviating
// the user's home directory as "~") and optionally the Git repository status
// with ANSI color sequences. The main function exposed is Update, which
// returns the formatted prompt string for the shell.
package prompt

import (
	"Ebash/internal/painter"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const DefaultPrompt = ">: "

// Update constructs and returns the prompt string for the shell. The prompt
// shows the current working directory, abbreviated with ~ for the user's
// home directory, and includes Git branch and status information. Paths
// deeper than three levels are shortened to ~/.../parent/child. Colors and
// bold styling are applied via the provided painter.Painter. If the current
// working directory or home directory cannot be determined, DefaultPrompt
// is returned.
func Update(painter painter.Painter) string {

	currPath, err := os.Getwd()
	if err != nil {
		return DefaultPrompt
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return DefaultPrompt
	}

	if home != "" && strings.HasPrefix(currPath, home) {
		currPath = "~" + strings.TrimPrefix(currPath, home)
	}

	currPathSplit := strings.Split(currPath, "/")
	if len(currPathSplit) > 3 {
		currPath = fmt.Sprintf("~/.../%s/%s", currPathSplit[len(currPathSplit)-2], currPathSplit[len(currPathSplit)-1])
	}

	pathStr := painter.Paint(painter.PathBold, painter.PathColour, currPath)
	gitStr := painter.Paint(painter.GitBold, painter.GitColour, gitStatus())

	return fmt.Sprintf("%s%s ", pathStr, gitStr)

}

// gitStatus returns a formatted string representing the Git branch and
// the current repository status. It shows the branch name and counts of
// modified and untracked files. Symbols used:
//
//	✓  - clean
//	✗  - modified
//	?  - untracked
//
// If the current directory is not a Git repository, an empty string is returned.
func gitStatus() string {

	branch, _ := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	branchStr := strings.TrimSpace(string(branch))

	if branchStr == "" {
		return ""
	}

	outStatus, _ := exec.Command("git", "status", "--porcelain").Output()

	lines := strings.Split(string(outStatus), "\n")

	var modified, untracked int

	for _, line := range lines {
		if strings.HasPrefix(line, " M") {
			modified++
		} else if strings.HasPrefix(line, "??") {
			untracked++
		}
	}

	switch {
	case modified == 0 && untracked == 0:
		return fmt.Sprintf("(%s ✓)", branchStr)
	case modified == 0 && untracked != 0:
		return fmt.Sprintf("(%s ?:%d ✗)", branchStr, untracked)
	case modified != 0 && untracked == 0:
		return fmt.Sprintf("(%s U:%d ✗)", branchStr, modified)
	default:
		return fmt.Sprintf("(%s U:%d ?:%d ✗)", branchStr, modified, untracked)
	}

}
