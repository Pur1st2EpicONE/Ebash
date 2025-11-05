![ebash banner](assets/banner.png)

<h3 align="center">Enhanced Bourne Again Shell — interactive Unix-like shell with pipelines, conditional execution, builtin commands, and external process handling.</h3>

<br>


## Table of Contents

1. [Architecture and key features](#architecture-and-key-features)
2. [Installation](#installation)
3. [Usage examples](#usage-examples)
4. [Configuration](#configuration)
5. [Running tests](#running-tests)
6. [Shutting down](#shutting-down)

<br>

## Architecture and key features

### Architecture

Ebash is organized around a small set of cohesive components designed to demonstrate how a minimal interactive shell can be built in Go. The main components are:

* **Shell** — the runtime orchestrator. It wires together the terminal (readline), prompt painter, completer, parser and command execution loop. It handles signal forwarding, lifecycle and descriptor leak checking.

* **Builtins** — synchronous implementations of common shell builtins (cd, pwd, echo, kill, ps) executed directly in the process.

* **External** — helpers that spawn and wait for external commands using os/exec, wiring stdin/stdout/stderr to support pipes and redirections.

* **Completer** — provides dynamic, context-aware tab completion by scanning the current directory and /proc, suggesting files, directories, and process IDs.

* **Parser** — lightweight parser that supports conditional operators (&&, ||), pipes (|) and simple redirections (<, >, >>).

* **Prompt / Painter** — builds a colored prompt that optionally includes compact git status (branch, modified/untracked counts) and path shortening.

At the top-level, Shell initializes configuration, the readline terminal and background goroutines (signal handler, completer updater) and then runs the interactive loop.

### Key features

#### Readline-based interactive experience

Built with [github.com/chzyer/readline](https://github.com/chzyer/readline) to provide line editing, history and prefix-based autocompletion.

#### Builtin and External commands

Ebash implements several builtins for educational purposes and delegates other commands to the OS via os/exec, preserving pipeline behaviour.

#### Tab completion

Completer dynamically recomputes suggestions on each loop: directory entries for file-oriented commands and process IDs for kill.

#### Command execution

Supports multiple chained pipes, simple input/output redirections, and conditional execution.

#### Git-aware, OhMyBash-inspired prompt

Displays a compact Git branch and status, abbreviates deep paths for readability, and shows relevant Git icons.

#### File descriptor leak detection

Shell records the baseline number of open file descriptors at startup and performs periodic checks. If a leak is detected the shell panics with details.

#### Simple, testable design

The project is intentionally small and oriented to tests that compare ebash output to the real bash for a number of scenarios.

<br>

## Installation

⚠️ Prerequisite: Go 1.25.1+

Clone the repository and enter the project directory:


```bash
git clone https://github.com/Pur1st2EpicONE/Ebash.git
cd Ebash
```

Optionally, edit [the config file](config.yaml) to customize your preferences, then build and run the project using the Makefile:

```bash
make
```

The make command checks for changes in the source files. If nothing has changed, it won’t rebuild the project. This means you can use make to launch Ebash even after exiting it, without triggering a rebuild.
