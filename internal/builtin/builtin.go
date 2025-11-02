// Package builtin implements a set of simple shell builtin commands used by ebash.
// It provides functions to execute builtins such as cd, pwd, echo, kill, and ps.
// The implementations are intentionally minimal and intended for educational
// purposes within the ebash project.
package builtin

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	ps "github.com/mitchellh/go-ps"
)

// Execute runs a builtin command based on the provided command slice.
// The function inspects command[0] and dispatches to the matching builtin
// implementation (cd, pwd, echo, kill, ps). If lastInPipeline is true,
// the output is directed to outputFile when it is non-nil, otherwise to stdout.
// Execute returns an error when a builtin reports failure, or nil on success.
func Execute(command []string, writer, outputFile *os.File, lastInPipeline bool) error {

	if lastInPipeline {
		if outputFile != nil {
			writer = outputFile
		} else {
			writer = os.Stdout
		}
	}

	switch command[0] {
	case "cd", "cd..":
		return changeDirectory(command)
	case "pwd":
		return printWorkingDirectory(writer)
	case "echo":
		return echo(command, writer)
	case "kill":
		return kill(command)
	case "ps":
		return processStatus(writer)
	}

	return nil

}

// changeDirectory changes the current working directory according to the
// arguments in the command slice. Returns an error for too many arguments
// or when the target path does not exist or is not a directory.
func changeDirectory(command []string) error {

	var dir string

	switch {
	case len(command) == 1 && command[0] == "cd..":
		dir = ".."
	case len(command) == 1 || command[1] == "~":
		dir = os.Getenv("HOME")
	case len(command) > 2:
		return fmt.Errorf("ebash: cd: too many arguments")
	default:
		dir = command[1]
	}

	if err := os.Chdir(dir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("ebash: cd: %s: Not a directory", dir)
		}
		return fmt.Errorf("ebash: cd: %w", err)
	}

	return nil

}

// printWorkingDirectory writes the current working directory path to the
// provided writer. Returns an error if the current directory cannot be
// determined.
func printWorkingDirectory(writer io.Writer) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("ebash: pwd: failed to get absolute path name: %w", err)
	}
	if _, err := fmt.Fprintln(writer, dir); err != nil {
		return fmt.Errorf("ebash: pwd: write operation failed: %w", err)
	}
	return nil
}

// echo prints the command arguments (excluding the command itself) to the
// provided writer, joined by spaces, followed by a newline.
func echo(command []string, writer io.Writer) error {
	if _, err := fmt.Fprintln(writer, strings.Join(command[1:], " ")); err != nil {
		return fmt.Errorf("ebash: echo: write operation failed: %w", err)
	}
	return nil
}

// kill sends SIGTERM to the process whose PID is specified by the first
// argument in command. Returns an error for invalid usage, non-numeric PID,
// or when the operation is not permitted.
func kill(command []string) error {

	if len(command) < 2 {
		return fmt.Errorf("kill: usage: kill [-s sigspec | -n signum | -sigspec] pid | jobspec ... or kill -l [sigspec]")
	}

	pid, err := strconv.Atoi(command[1])
	if err != nil {
		return fmt.Errorf("ebash: kill: %s: arguments must be process or job IDs", command[1])
	}
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("ebash: kill: (%d) - Operation not permitted", pid)
	}

	return nil

}

// processStatus prints a simple ps-like listing of processes that are
// attached to the same terminal as the current process. It relies on psPrep
// to obtain the terminal path, a matching regexp, and the list of processes.
func processStatus(writer io.Writer) error {

	path, re, processes, err := psPrep(writer)
	if err != nil {
		return fmt.Errorf("ebash: ps: %w", err)
	}

	var pid int
	var cmd string

	for _, process := range processes {

		pid = process.Pid()
		cmd = process.Executable()

		link, err := os.Readlink(fmt.Sprintf("/proc/%d/fd/0", pid))
		if err == nil && re.MatchString(link) {
			if _, err = fmt.Fprintf(writer, "%7d pts/%-8s 00:00:00 %s\n", pid, filepath.Base(path), cmd); err != nil {
				return fmt.Errorf("write operation failed: %w", err)
			}
		}

	}

	return nil

}

// psPrep prepares and returns data needed by processStatus: the path to the
// current terminal, a compiled regexp that matches that terminal, the list
// of processes, and an error if any step fails.
func psPrep(writer io.Writer) (string, *regexp.Regexp, []ps.Process, error) {

	path, err := os.Readlink("/proc/self/fd/0")
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to read /proc/self/fd/0: %w", err)
	}

	re := regexp.MustCompile(fmt.Sprintf(`/dev/pts/%s$`, filepath.Base(path)))

	processes, err := ps.Processes()
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to get process list: %w", err)
	}

	if _, err := fmt.Fprintln(writer, "    PID TTY          TIME CMD"); err != nil {
		return "", nil, nil, fmt.Errorf("write operation failed: %w", err)
	}
	return path, re, processes, nil

}
