// Package prompt provides a small utility to build the interactive shell
// prompt string. It renders the current working directory (using ~ for the
// user's home directory) with ANSI color escapes and exposes a single
// Update function used by the shell to obtain the prompt.
package prompt

import (
	"os"
	"strings"
)

const (
	green         = "\033[32m"
	blue          = "\033[94m"
	reset         = "\033[0m"
	DefaultPrompt = "$ "
)

// Update returns the prompt string to be displayed to the user. The prompt
// shows the current working directory (with the home directory abbreviated
// as `~` when applicable) wrapped in ANSI color sequences. If the working
// directory cannot be determined, DefaultPrompt is returned.
func Update() string {

	currPath, err := os.Getwd()
	if err != nil {
		return DefaultPrompt
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = ""
	}

	promptPath := currPath
	if homeDir != "" && strings.HasPrefix(currPath, homeDir) {
		promptPath = "~" + strings.TrimPrefix(currPath, homeDir)
	}

	return green + promptPath + blue + reset + " "

}
