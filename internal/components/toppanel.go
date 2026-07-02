package components

import (
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
)

type Top struct {
	trackCount int
}

var (
	bullet = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))
	
	topStyle = lipgloss.NewStyle().
		Bold(true).
		Italic(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color(""))
)

func (t *Top) SetTrackCount(n int) {
	t.trackCount = n
}

func (t *Top) View(height, width int) string {
	now := time.Now()
	bullet := bullet.Render(" ▐ ")
	left := bullet + "Resonance  ["
	if t.trackCount == 1 {
		left += "1 track"
	} else {
		left += fmt.Sprintf("%d tracks", t.trackCount)
	}
	left += "]"
	right := now.Format("  Monday, Jan 2  3:04 PM  ")

	space := width - lipgloss.Width(left) - lipgloss.Width(right)
	if space < 1 {
		space = 1
	}

	fill := ""
	for i := 0; i < space; i++ {
		fill += " "
	}

	return topStyle.Height(height).Width(width).Render(left + fill + right)
}
