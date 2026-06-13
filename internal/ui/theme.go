package ui

import "github.com/charmbracelet/lipgloss"

var (
	styleDim      = lipgloss.NewStyle().Faint(true)
	styleBoldCyan = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#0070f3", Dark: "#00b4d8"})
	styleLabel    = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#888888"})
	styleAccent   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0070f3", Dark: "#00b4d8"})
	styleValue    = lipgloss.NewStyle().Bold(true)
	styleSep      = lipgloss.NewStyle().Faint(true)
	styleOK       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#16a34a", Dark: "#4ade80"})
	styleErr      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#dc2626", Dark: "#f87171"})
	styleSnippet  = lipgloss.NewStyle().Faint(true)
)
