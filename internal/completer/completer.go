// Package completer provides filesystem- and process-aware tab completion
// for the ebash shell. It generates suggestions for commands like `cd`,
// `kill`, `ls`, `cat`, `vim`, and others based on the current directory
// contents and currently running process IDs.
package completer

import (
	"os"
	"strconv"

	"github.com/chzyer/readline"
)

// Update scans the current working directory and running processes
// to build a new readline.AutoCompleter instance for ebash. It provides:
//
//   - Directory suggestions for `cd`.
//   - Process IDs for `kill`.
//   - File and directory names for `ls`, `cat`, `vim`, `cut`, `grep`, `echo`, etc.
//   - File and directory names for `rm` and `rm -rf` (supports both plain `rm` and `rm -rf <dir>`).
//
// Returns the configured AutoCompleter, or nil if the current directory
// cannot be read.
func Update() readline.AutoCompleter {

	entries, err := os.ReadDir(".")
	if err != nil {
		return nil
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

	completer := readline.NewPrefixCompleter(
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

	return completer

}

// getPIDs reads the /proc directory to find all currently running
// process IDs. It returns a slice of PID strings, which is used
// to provide completion suggestions for the `kill` command.
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
