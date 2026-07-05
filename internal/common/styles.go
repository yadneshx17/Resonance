package common

import (
	"charm.land/lipgloss/v2"
)

var (
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	PlayingIconStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00FF00"))

	PausedIconStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700"))

	PlayingTrackStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#20002F")).
				Background(lipgloss.Color("#F5F5DC"))

	CursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD700"))
)
