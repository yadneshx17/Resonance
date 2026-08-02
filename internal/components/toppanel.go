package components

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

type Top struct {
	trackCount int
	source     string
}

func (t *Top) SetTrackCount(n int, source string) {
	t.trackCount = n
	t.source = source
}

func (t *Top) View(height, width int) string {
	left := "  Resonance   "

	if t.source != "spotify" {
		if t.trackCount == 1 {
			left += "1 track"
		} else {
			left += fmt.Sprintf(" ‣ %d tracks", t.trackCount)
		}
	} else {
		left += fmt.Sprint("‣" + " Spotify")
	}

	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7086")).
		Render(" <?> help  ")

	space := width - (lipgloss.Width(left) + lipgloss.Width(hint))
	if space < 1 {
		space = 1
	}

	fill := ""
	for i := 0; i < space; i++ {
		fill += " "
	}

	style := lipgloss.NewStyle().
		Bold(true).
		Italic(true).
		Foreground(lipgloss.Color("#CDD6F4")).
		Height(height).
		Width(width)
	return style.Render(left + fill + hint)
}
