package components

import (
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
)

type Top struct {
	trackCount int
}

func (t *Top) SetTrackCount(n int) {
	t.trackCount = n
}

func (t *Top) View(height, width int) string {
	now := time.Now()

	left := "  Resonance   "
	if t.trackCount == 1 {
		left += "1 track"
	} else {
		left += fmt.Sprintf(" %d tracks", t.trackCount)
	}
	right := now.Format("  Monday, Jan 2  3:04 PM  ")

	space := width - (lipgloss.Width(left) + lipgloss.Width(right))
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
		Foreground(lipgloss.Color("#ffffff")).
		Height(height).
		Width(width)
	return style.Render(left + fill + right)
}
