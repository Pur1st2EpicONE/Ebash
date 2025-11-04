// Package ebash contains the core interactive shell loop and orchestration
// logic for the ebash project. It wires together configuration, the
// readline-based terminal, builtin command execution, external command
// execution, and signal handling.
package ebash

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"

	"github.com/chzyer/readline"

	"Ebash/internal/builtin"
	"Ebash/internal/completer"
	"Ebash/internal/config"
	"Ebash/internal/external"
	"Ebash/internal/painter"
	"Ebash/internal/parser"

	"Ebash/internal/prompt"
)

// Shell holds the runtime state of the interactive shell. It contains
// synchronization primitives, channels for signal handling and shutdown,
// the parsed pipeline for the current input line, the readline terminal
// instance, a set of supported builtins, and a list of currently running
// external commands.
type Shell struct {
	mu            sync.Mutex          // protects mutable fields (e.g. externals)
	sigCh         chan os.Signal      // receives OS signals (e.g. os.Interrupt)
	stopCh        chan struct{}       // closed to request shutdown of background goroutines
	painter       painter.Painter     // renders the shell prompt with colors and styles
	pipeline      []parser.Pipe       // parsed pipeline: sequence of conditional Pipe sections
	terminal      *readline.Instance  // readline instance used to read user input
	builtins      map[string]struct{} // set of builtin command names for quick lookup
	externals     []*exec.Cmd         // running external commands tracked for signaling/waiting
	descriptors   int                 // baseline number of file descriptors at shell startup
	checkCounter  uint                // incremented each pipeline; fd check runs only when reaching checkInterval
	checkInterval uint                // number of pipelines between descriptor checks; set to 0 in config to disable
}

// Run starts the main interactive loop of the shell. It boots the shell,
// then repeatedly reads lines from the terminal, parses them into pipelines,
// executes those pipelines and reports any errors. The function returns only
// when EOF is received or the user executes the "exit" command.
func Run() {

	shell, err := boot()
	if err != nil {
		panic(err)
	}

	defer shell.exit()

	for {

		shell.terminal.Config.AutoComplete = completer.Update()
		shell.terminal.SetPrompt(prompt.Update(shell.painter))

		line, err := shell.terminal.Readline()
		if err != nil {
			if errors.Is(err, readline.ErrInterrupt) {
				continue
			} else if errors.Is(err, io.EOF) {
				return
			}
			panic(err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		} else if line == "exit" {
			fmt.Println(line)
			return
		}

		shell.pipeline, err = parser.Parse(line)
		if err != nil {
			shell.sysmon(err)
			continue
		}

		shell.sysmon(shell.runPipeline())

	}

}

// boot initializes the shell runtime. It loads configuration (falling back
// to defaults on error), creates a readline terminal instance, records the
// baseline number of file descriptors for later leak detection, sets up the
// builtin command table, initializes the prompt painter, and starts the
// interrupt handler goroutine.
// Returns the initialized Shell or an error if initialization fails.
func boot() (*Shell, error) {

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		cfg = config.Default()
	}

	readlineCfg := &readline.Config{
		HistoryFile:     cfg.Terminal.HistoryFile,
		HistoryLimit:    cfg.Terminal.HistoryLimit,
		InterruptPrompt: cfg.Terminal.InterruptPrompt,
		EOFPrompt:       "\n" + cfg.Terminal.EOFPrompt,
	}

	terminal, err := readline.NewEx(readlineCfg)
	if err != nil {
		return nil, fmt.Errorf("ebash: boot: failed to create new terminal instance: %w", err)
	}

	descriptors, err := os.ReadDir(fmt.Sprintf("/proc/%d/fd", os.Getpid()))
	if err != nil {
		return nil, fmt.Errorf("ebash: boot: cannot read fd directory: %w", err)
	}

	shell := &Shell{
		terminal:      terminal,
		sigCh:         make(chan os.Signal, 1),
		stopCh:        make(chan struct{}),
		descriptors:   len(descriptors),
		checkInterval: cfg.Terminal.CheckInterval,
		painter:       painter.NewPainter(cfg.Prompt),
		builtins: map[string]struct{}{
			"cd":   {},
			"cd..": {},
			"pwd":  {},
			"echo": {},
			"kill": {},
			"ps":   {},
		},
	}

	signal.Notify(shell.sigCh, os.Interrupt)
	go shell.interruptHandler()

	return shell, nil

}

// interruptHandler listens for OS interrupt signals (SIGINT) and forwards
// them as Interrupt signals to any running external commands. The goroutine
// exits when the shell stop channel is closed.
func (shell *Shell) interruptHandler() {
	for {
		select {
		case <-shell.stopCh:
			return
		case <-shell.sigCh:
			shell.mu.Lock()
			for _, externalCommand := range shell.externals {
				_ = externalCommand.Process.Signal(os.Interrupt) // https://www.youtube.com/watch?v=g3m369iaOlI
			}
			shell.mu.Unlock()
		}
	}
}

// exit performs cleanup of the shell runtime: it stops signal delivery,
// signals the interrupt handler to stop, and closes the readline terminal.
func (shell *Shell) exit() {
	signal.Stop(shell.sigCh)
	close(shell.stopCh)
	_ = shell.terminal.Close()
}

// runPipeline executes the parsed pipeline (which may contain multiple pipe
// segments). It honors conditional execution flags (NextAnd/NextOr) between
// pipeline segments and returns the first error encountered, if any.
func (shell *Shell) runPipeline() error {

	var shouldRun bool
	var lastExitCode int

	for i := 0; i < len(shell.pipeline); i++ {

		pipe := shell.pipeline[i]
		shouldRun = true

		if i > 0 {

			previousPipe := shell.pipeline[i-1]

			if previousPipe.NextAnd && lastExitCode != 0 {
				shouldRun = false
			} else if previousPipe.NextOr && lastExitCode == 0 {
				shouldRun = false
			}

		}

		if shouldRun {
			exitCode, err := shell.runPipe(pipe)
			lastExitCode = exitCode
			if err != nil {
				return err
			}
		}

	}

	return nil

}

// runPipe executes a single pipe segment composed of multiple commands
// connected by pipes. Builtin commands are executed synchronously via the
// builtin package; external commands are spawned and tracked. The function
// wires up pipes between commands, handles input/output redirection, waits
// for external processes to finish, and returns the exit code and an error
// if any operation fails.
func (shell *Shell) runPipe(pipe parser.Pipe) (int, error) {

	var err error
	var lastInSection bool
	var writer, connector, reader *os.File

	for i, command := range pipe.Section {

		lastInSection = (i == len(pipe.Section)-1)

		if !lastInSection {
			reader, writer, err = os.Pipe()
			if err != nil {
				closeDescriptors(writer, connector, reader)
				return 1, err
			}
		}

		if _, builtinCommand := shell.builtins[command[0]]; builtinCommand {
			err = builtin.Execute(command, writer, pipe.Output, lastInSection)
		} else {
			execCmd, externalError := external.Execute(command, writer, connector, pipe.Input, pipe.Output, lastInSection)
			if externalError == nil {
				shell.mu.Lock()
				shell.externals = append(shell.externals, execCmd)
				shell.mu.Unlock()
			} else {
				err = externalError
			}

		}

		if err != nil {
			closeDescriptors(writer, connector, reader, pipe.Input, pipe.Output)
			return 1, err
		}

		closeDescriptors(writer, connector)

		if !lastInSection {
			connector = reader
		}

	}

	closeDescriptors(reader, pipe.Input, pipe.Output)

	if shell.externals != nil {
		err = shell.sync()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				return exitErr.ExitCode(), nil
			}
			return 1, err
		}
	}

	return 0, nil

}

// closeDescriptors closes each provided *os.File descriptor if it is non-nil
// and not one of the standard input/output descriptors. This is a helper used
// to ensure pipes and temporary files are properly closed.
func closeDescriptors(descriptors ...*os.File) {
	for _, descriptor := range descriptors {
		if descriptor != nil && descriptor != os.Stdin && descriptor != os.Stdout {
			_ = descriptor.Close()
		}
	}
}

// sync waits for any tracked external commands to finish
// and resets the external command list. It returns any
// error returned by external.Wait.
func (shell *Shell) sync() error {

	shell.mu.Lock()

	err := external.Wait(shell.externals)
	shell.externals = nil

	shell.mu.Unlock()

	return err

}

// sysmon monitors the shell’s runtime state. It logs any provided errors
// and checks for file descriptor leaks relative to the baseline count.
// The check is performed only every `checkInterval` pipelines; `checkCounter`
// is incremented on each pipeline execution and reset after the check.
// If more descriptors are open than the baseline, the function panics
// and reports the PID along with the currently open file descriptors.
func (shell *Shell) sysmon(err error) {

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	shell.checkCounter++

	if shell.checkCounter == shell.checkInterval && shell.checkInterval != 0 {

		pid := os.Getpid()
		fdDir := fmt.Sprintf("/proc/%d/fd", pid)
		currDescriptors, err := os.ReadDir(fdDir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "sysmon: cannot read fd dir:", err)
			return
		}

		if len(currDescriptors) > shell.descriptors {

			openDescriptors := []string{}
			for _, openDescriptor := range currDescriptors {
				openDescriptors = append(openDescriptors, openDescriptor.Name())
			}

			panic(fmt.Errorf(
				"descriptor leak detected: %d file descriptors still open (PID=%d, open fds=%v)",
				len(currDescriptors)-shell.descriptors,
				os.Getpid(),
				openDescriptors,
			))

		}

		shell.checkCounter = 0

	}

}
