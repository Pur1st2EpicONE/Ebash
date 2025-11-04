// Package external provides helpers to spawn and wait for external commands
// executed by the ebash shell. It wraps os/exec to set up stdin/stdout/stderr
// based on pipeline connectors and redirection files.
package external

import (
	"os"
	"os/exec"
)

// Execute starts an external command described by the command slice.
// It configures standard input and output depending on the provided
// connector (previous pipe), inputFile/outputFile (redirection), and
// whether this command is the last in the pipeline.
//
// Special handling: if the command is "ls" or "grep" and the output
// is not redirected to a file, the "--color=always" flag is prepended
// to force color output.
//
// Returns the started *exec.Cmd so the caller can track/wait on it,
// or an error if starting the process fails.
func Execute(command []string, writer, connector, inputFile, outputFile *os.File, isLast bool) (*exec.Cmd, error) {

	args := command[1:]
	if command[0] == "ls" || command[0] == "grep" {
		if outputFile == nil {
			args = append([]string{"--color=always"}, args...)
		}
	}

	cmd := exec.Command(command[0], args...)
	cmd.Stderr = os.Stderr

	if connector != nil {
		cmd.Stdin = connector
	} else if inputFile != nil {
		cmd.Stdin = inputFile
	} else {
		cmd.Stdin = os.Stdin
	}

	if !isLast {
		cmd.Stdout = writer
	} else if outputFile != nil {
		cmd.Stdout = outputFile
	} else {
		cmd.Stdout = os.Stdout
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return cmd, nil
}

// Wait blocks until all provided external commands have finished.
// It returns the last non-nil error observed, or nil if all commands
// exited successfully. This mirrors pipeline behavior: all processes
// are waited on, but the last error is returned for reporting.
func Wait(externals []*exec.Cmd) error {
	var lastErr error
	for _, command := range externals {
		if err := command.Wait(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
