// Package completer provides filesystem- and process-aware tab completion
// for the ebash shell. It dynamically builds completion suggestions for
// common shell commands based on the current directory contents and running
// system processes.
package completer

import (
	"os"
	"strconv"

	"github.com/chzyer/readline"
)

// Completer adapts ebash's dynamic environment (filesystem and processes)
// to the readline.AutoCompleter interface. It generates and updates
// command-specific completion suggestions on each loop iteration.
type Completer struct {
	readlineCompleter *readline.PrefixCompleter
}

// NewCompleter returns a new Completer instance with an empty
// underlying PrefixCompleter.
func NewCompleter() *Completer {
	return &Completer{readlineCompleter: readline.NewPrefixCompleter()}
}

// Update rebuilds the completion tree based on the current working directory
// and system state. It scans files, directories, and running processes to
// provide up-to-date suggestions for commands like "cd", "ls", "kill",
// "rm", "cat", and others.
func (c *Completer) Update() {

	entries, err := os.ReadDir(".")
	if err != nil {
		return
	}

	var onlyDirs []readline.PrefixCompleterInterface
	var procsToKill []readline.PrefixCompleterInterface
	var rmCompleter []readline.PrefixCompleterInterface
	var fileNamesToComplete []readline.PrefixCompleterInterface

	for _, entry := range entries {
		if entry.IsDir() {
			fileNamesToComplete = append(fileNamesToComplete, readline.PcItem(entry.Name()+"/"))
			onlyDirs = append(onlyDirs, readline.PcItem(entry.Name()+"/"))
		} else {
			fileNamesToComplete = append(fileNamesToComplete, readline.PcItem(entry.Name()))
		}
	}

	toKill := getPIDs()
	for _, val := range toKill {
		procsToKill = append(procsToKill, readline.PcItem(val))
	}

	rmCompleter = append(rmCompleter, fileNamesToComplete...)
	rmCompleter = append(rmCompleter, readline.PcItem("-rf", fileNamesToComplete...))

	newCompleter := readline.NewPrefixCompleter(
		readline.PcItem("cd", onlyDirs...),
		readline.PcItem("rm", rmCompleter...),
		readline.PcItem("kill", procsToKill...),
		readline.PcItem("ps", fileNamesToComplete...),
		readline.PcItem("ls", fileNamesToComplete...),
		readline.PcItem("cat", fileNamesToComplete...),
		readline.PcItem("cut", fileNamesToComplete...),
		readline.PcItem("vim", fileNamesToComplete...),
		readline.PcItem("grep", fileNamesToComplete...),
		readline.PcItem("echo", fileNamesToComplete...),
	)

	c.readlineCompleter = newCompleter

}

// Do delegates the completion logic to the underlying PrefixCompleter.
// It satisfies the readline.AutoCompleter interface.
func (c *Completer) Do(line []rune, pos int) ([][]rune, int) {
	return c.readlineCompleter.Do(line, pos)
}

// getPIDs reads the /proc directory to find all currently running
// process IDs. It returns a slice of PID strings, which is used
// to provide completion suggestions for the "kill" command.
func getPIDs() []string {
	proc, _ := os.ReadDir("/proc")
	var pids []string
	for _, entry := range proc {
		if entry.IsDir() {
			name := entry.Name()
			if _, err := strconv.Atoi(name); err == nil {
				pids = append(pids, name)
			}
		}
	}
	return pids
}
