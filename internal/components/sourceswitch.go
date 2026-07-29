package components

import (
	"strings"

	"charm.land/lipgloss/v2"
)

type SourceSwitchData struct {
	Visible      bool
	SourceCursor int  // 0=Local, 1=Spotify
	SpotifyOK    bool // logged in
	Width        int
	Height       int
}

type SourceSwitch struct {
	visible      bool
	sourceCursor int
	spotifyOK    bool
	width        int
	height       int
}

func (s *SourceSwitch) SetData(data SourceSwitchData) {
	s.visible = data.Visible
	s.sourceCursor = data.SourceCursor
	s.spotifyOK = data.SpotifyOK
	s.width = data.Width
	s.height = data.Height
}

func (s SourceSwitch) View() string {
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A6E3A1"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CBA6F7"))

	items := []string{}
	for i, label := range []string{"Local", "Spotify"} {
		prefix := "  "
		style := dimStyle
		if i == s.sourceCursor {
			prefix = "> "
			style = selectedStyle
		}
		suffix := ""
		if i == 1 && !s.spotifyOK {
			suffix = dimStyle.Render(" (not logged in)")
		}
		items = append(items, "  "+style.Render(prefix+label)+suffix)
	}

	content := titleStyle.Render("Source") + "\n\n" + strings.Join(items, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#A6E3A1")).
		Padding(1, 3).
		Width(30).
		Render(content)

	return box
}
