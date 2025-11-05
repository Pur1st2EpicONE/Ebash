// Package config provides functionality for loading shell configuration
// parameters from a config file using the Viper library. It defines terminal
// behavior and prompt appearance settings.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all configurable settings for the shell, including
// terminal behavior and prompt appearance.
type Config struct {
	Terminal Terminal `mapstructure:"terminal"` // Terminal-related settings
	Prompt   Prompt   `mapstructure:"prompt"`   // Prompt appearance settings
}

// Terminal defines settings related to terminal behavior, such as history
// file, history limit, interrupt and exit prompts, and file descriptor check interval.
type Terminal struct {
	HistoryFile     string `mapstructure:"history_file"`     // Path to shell history file
	HistoryLimit    int    `mapstructure:"history_limit"`    // Maximum number of history entries
	InterruptPrompt string `mapstructure:"interrupt_prompt"` // Text shown on Ctrl-C
	EOFPrompt       string `mapstructure:"exit_message"`     // Text shown on EOF/exit
	CheckInterval   uint   `mapstructure:"check_interval"`   // Number of pipelines between FD checks
}

// Prompt defines settings related to the shell prompt appearance,
// including theme, colors, and bold styling.
type Prompt struct {
	Theme               string `mapstructure:"theme"`                  // Prompt theme name
	PathColour          string `mapstructure:"path_colour"`            // Color for current path
	PathColourBold      bool   `mapstructure:"path_colour_bold"`       // Bold style for path
	GitStatusColour     string `mapstructure:"git_status_colour"`      // Color for git branch/status
	GitStatusColourBold bool   `mapstructure:"git_status_colour_bold"` // Bold style for git info
}

// Load reads configuration from a file named "config" in the current
// directory using Viper, and unmarshals it into a Config instance. Returns
// a partial Config and an error if loading or unmarshaling fails.
func Load() (*Config, error) {

	viper.AddConfigPath(".")
	viper.SetConfigName("config")

	cfg := new(Config)

	if err := viper.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("failed to load config: %v", err)
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return cfg, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return cfg, nil
}

// Default returns a Config with sensible default settings. It is used
// as a fallback when loading a configuration file fails.
func Default() *Config {

	cfg := new(Config)

	cfg.Terminal.HistoryFile = filepath.Join(os.Getenv("HOME"), ".ebash_history")
	cfg.Terminal.HistoryLimit = 1000
	cfg.Terminal.InterruptPrompt = "^C"
	cfg.Terminal.EOFPrompt = "exit"
	cfg.Terminal.CheckInterval = 5

	cfg.Prompt.Theme = "default"
	cfg.Prompt.PathColour = "\033[32m"
	cfg.Prompt.PathColourBold = false
	cfg.Prompt.GitStatusColour = "\033[94m"
	cfg.Prompt.GitStatusColourBold = true

	return cfg
}
