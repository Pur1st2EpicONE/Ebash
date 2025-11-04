// Package painter provides functionality to render colored and styled
// text for the shell prompt. It supports path and Git status coloring
// with optional bold formatting and can apply pre-defined themes.
package painter

import (
	"Ebash/internal/config"
	"strings"
)

const (
	reset    = "\033[0m"
	makeBold = "\033[1m"
)

// Painter holds styling information for the shell prompt, including
// path and Git colors and bold settings.
type Painter struct {
	PathColour string // ANSI or Unicode escape code for the path
	PathBold   bool   // Whether the path should be bold
	GitColour  string // ANSI or Unicode escape code for Git info
	GitBold    bool   // Whether Git info should be bold
}

// NewPainter creates a new Painter based on the provided config.Prompt.
// If a theme is set in the config, it will override the colors below.
// Otherwise, colors are taken directly from the config fields.
func NewPainter(cfg config.Prompt) Painter {
	if cfg.Theme != strings.TrimSpace("none") || cfg.Theme != strings.TrimSpace("") {
		resolveTheme(&cfg)
	}
	return Painter{
		PathColour: resolveColor(cfg.PathColour),
		PathBold:   cfg.PathColourBold,
		GitColour:  resolveColor(cfg.GitStatusColour),
		GitBold:    cfg.GitStatusColourBold,
	}
}

// resolveTheme applies a predefined theme to the provided Prompt config.
func resolveTheme(cfg *config.Prompt) {

	theme := strings.TrimSpace(cfg.Theme)
	if theme == "" {
		return
	}

	switch strings.ToLower(theme) {
	case "ebash":
		setEbash(cfg)
	case "wildberries":
		setWildberries(cfg)
	case "monokai":
		setMonokai(cfg)
	case "ohmybash":
		setOhMyBash(cfg)
	}

}

// setEbash applies the default ebash theme.
func setEbash(cfg *config.Prompt) {

	cfg.PathColour = "yellow"
	cfg.PathColourBold = false
	cfg.GitStatusColour = "default"
	cfg.GitStatusColourBold = false

}

// setWildberries applies the Wildberries theme.
func setWildberries(cfg *config.Prompt) {

	cfg.PathColour = "\u001b[38;2;203;17;171m"
	cfg.PathColourBold = true
	cfg.GitStatusColour = "default"
	cfg.GitStatusColourBold = true

}

// setMonokai applies the Monokai theme.
func setMonokai(cfg *config.Prompt) {

	cfg.PathColour = "\u001b[38;2;249;38;114m"
	cfg.PathColourBold = true
	cfg.GitStatusColour = "\u001b[38;2;166;226;46m"
	cfg.GitStatusColourBold = false

}

// setOhMyBash applies the OhMyBash theme.
func setOhMyBash(cfg *config.Prompt) {

	cfg.PathColour = "green"
	cfg.PathColourBold = false
	cfg.GitStatusColour = "blue"
	cfg.GitStatusColourBold = true

}

// resolveColor converts a color name or escape sequence string into
// a valid ANSI/Unicode escape code. If the input is already an escape
// sequence, it is returned unchanged.
func resolveColor(colour string) string {

	colour = strings.TrimSpace(colour)
	if colour == "" {
		return ""
	}

	switch strings.ToLower(colour) {
	case "default":
		return "\u001b[39m"
	case "black":
		return "\033[30m"
	case "red":
		return "\033[31m"
	case "green":
		return "\033[32m"
	case "yellow":
		return "\033[33m"
	case "bright yellow":
		return "\u001b[93m"
	case "blue":
		return "\033[94m"
	case "magenta":
		return "\033[35m"
	case "cyan":
		return "\033[36m"
	case "white":
		return "\033[37m"
	default:
		return colour
	}

}

// Paint applies the provided bold and color settings to the given text
// and returns the formatted string with ANSI escape sequences.
func (p Painter) Paint(bold bool, colour string, text string) string {
	style := ""
	if bold {
		style = makeBold
	}
	return style + colour + text + reset
}
