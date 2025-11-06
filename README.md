![ebash banner](assets/banner.png)

<h3 align="center">Enhanced Bourne Again Shell — interactive Unix-like shell with pipelines, conditional execution, builtin commands, and external process handling.</h3>

<br>

## Architecture

Ebash is organized around a small set of cohesive components designed to demonstrate how a minimal interactive shell can be built in Go. The main components are:

* **Shell** — the runtime orchestrator. It wires together the terminal (readline), prompt painter, completer, parser and command execution loop. It handles signal forwarding, lifecycle and descriptor leak checking.

* **Parser** — lightweight parser that supports conditional operators (&&, ||), pipes (|) and simple redirections (<, >, >>).

* **Builtins** — synchronous implementations of common shell builtins (cd, pwd, echo, kill, ps) executed directly in the process.

* **Externals** — helpers that spawn and wait for external commands using os/exec, wiring stdin/stdout/stderr to support pipes and redirections.

* **Completer** — provides dynamic, context-aware tab completion by scanning the current directory and /proc, suggesting files, directories, and process IDs.

* **Prompt / Painter** — builds a colored prompt that optionally includes compact git status (branch, modified/untracked counts) and path shortening.

<br>

## Cool features

#### Command execution

Ebash implements several builtins for educational purposes and executes both internal and external commands (via os/exec), supporting multiple chained pipes, simple input/output redirections, and conditional execution while preserving native Bash-like pipeline behavior.

#### Readline-based interactive experience

Built with [github.com/chzyer/readline](https://github.com/chzyer/readline) to provide line editing, command history, and prefix-based autocompletion. Ebash extends this with a custom completer that dynamically recomputes suggestions on each loop — offering directory entries for file-oriented commands and process IDs for kill.

#### Git-aware, OhMyBash-inspired prompt

Displays a compact Git branch and status, abbreviates deep paths for readability, and shows relevant Git icons. The prompt is fully customizable via the configuration file.

#### File descriptor leak detection

Shell records the baseline number of open file descriptors at startup and performs periodic checks. If a leak is detected the shell panics with details.

<br>

## Installation

⚠️ Prerequisite: Go 1.25.1

Clone the repository and enter the project directory:


```bash
git clone https://github.com/Pur1st2EpicONE/Ebash.git
cd Ebash
```

Optionally, edit [the config file](config.yaml) to customize your preferences, then build and run the project using the Makefile command:

```bash
make
```

<br>

## Testing & Linting

Run tests and ensure code quality:

```bash
make test        # Compare output with Bash (tested on Linux; results may differ on macOS)
make lint        # Linting checks
```