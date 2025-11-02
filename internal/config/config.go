// Package config provides functionality for loading configuration
// parameters from a config file using the Viper library.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds user-configurable settings for the shell.
type Config struct {
	HistoryFile     string `mapstructure:"history_file"`
	HistoryLimit    int    `mapstructure:"history_limit"`
	InterruptPrompt string `mapstructure:"interrupt_prompt"`
	EOFPrompt       string `mapstructure:"exit_message"`
}

// Load reads configuration from a file named "config" in the current
// directory using Viper and unmarshals it into a Config instance. If
// reading or unmarshaling fails an error is returned along with a partial
// Config (which may be zero-valued).
func Load() (*Config, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	cfg := new(Config)
	if err := viper.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("ebash: boot: failed to load config: %v", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("ebash: boot: failed to unmarshal config: %v", err)
	}
	return cfg, nil
}

// Default returns a Config populated with sensible defaults. This is used
// as a fallback when loading the configuration file fails.
func Default() *Config {
	return &Config{
		HistoryFile:     filepath.Join(os.Getenv("HOME"), ".ebash_history"),
		HistoryLimit:    1000,
		InterruptPrompt: "^C",
		EOFPrompt:       "\nexit",
	}
}
