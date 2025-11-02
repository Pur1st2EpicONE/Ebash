// Package external provides helpers to spawn and wait for external commands
// executed by the ebash shell. It wraps os/exec to set up stdin/stdout/stderr
// based on pipeline connectors and redirection files.
package external

import (
	"os"
	"os/exec"
)

// Execute starts an external command described by the command slice. The
// function configures the process's standard input and output depending on
// the provided connector (previous pipe), inputFile (redirection), outputFile
// (redirection) and whether this command is the last in the pipeline.
// Returns the started *exec.Cmd (so the caller can track/wait on it) or an
// error if starting the process fails.
func Execute(command []string, writer, connector, inputFile, outputFile *os.File, isLast bool) (*exec.Cmd, error) {

	cmd := exec.Command(command[0], command[1:]...)
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

// Wait blocks until all provided external commands have finished. It returns
// the last non-nil error observed from waits, or nil if all commands exited
// successfully. This mirrors the behavior of waiting for a pipeline of
// processes where earlier errors are recorded but the pipeline continues to
// be waited on.
func Wait(externals []*exec.Cmd) error {
	var lastErr error
	for _, command := range externals {
		if err := command.Wait(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
