package ui

import (
	"os"

	"golang.org/x/term"
)

// Mode is the selected output rendering.
type Mode int

// Output rendering modes, in priority order (JSON > HUD > Plain).
const (
	ModeHUD   Mode = iota // colored pipeline, TTY only
	ModePlain             // aligned text, no ANSI
	ModeJSON              // machine-readable, locked schema
)

// Env captures the inputs to output-mode selection.
type Env struct {
	IsTTY   bool
	NoColor bool
	JSON    bool
}

// SelectMode picks the renderer: JSON wins if requested; HUD only on a color TTY.
func SelectMode(e Env) Mode {
	switch {
	case e.JSON:
		return ModeJSON
	case e.IsTTY && !e.NoColor:
		return ModeHUD
	default:
		return ModePlain
	}
}

// DetectEnv inspects the real process: stdout TTY-ness and NO_COLOR.
func DetectEnv(json bool) Env {
	return Env{
		IsTTY:   term.IsTerminal(int(os.Stdout.Fd())),
		NoColor: os.Getenv("NO_COLOR") != "",
		JSON:    json,
	}
}
