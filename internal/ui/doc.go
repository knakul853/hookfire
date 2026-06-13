// Package ui renders the outcome of a trigger/replay as one of three views: a
// colored HUD pipeline on a TTY, aligned plain text off-TTY, or a locked JSON
// schema. Color + HUD apply only when stdout is a TTY, --json is off, and
// NO_COLOR is unset; everything degrades cleanly otherwise. The signing secret
// never appears in any view.
package ui
