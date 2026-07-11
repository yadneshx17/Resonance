package common

import (
	"charm.land/lipgloss/v2"
)

var (
	PlayingIconStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#A6E3A1"))

	PausedIconStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F5E0DC"))

	PlayingTrackStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#A6E3A1")).
				Background(lipgloss.Color("#1E1E2E"))

	CursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#CBA6F7"))
)
