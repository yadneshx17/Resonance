package components

import (
	"time"

	"charm.land/lipgloss/v2"
)

type VisualData struct {
	TrackName string
	StateIcon string
	Position  time.Duration
	Duration  time.Duration
	VolLevel  float64
	Muted     bool
	Height    int
	Width     int
}

type Visual struct {
	trackName string
	stateIcon string
	position  time.Duration
	duration  time.Duration
	volLevel  float64
	muted     bool
	height    int
	width     int
}

var visualStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder())

func (v *Visual) SetData(data VisualData) {
	v.trackName = data.TrackName
	v.stateIcon = data.StateIcon
	v.position = data.Position
	v.duration = data.Duration
	v.volLevel = data.VolLevel
	v.muted = data.Muted
	v.height = data.Height
	v.width = data.Width
}

func (v *Visual) View() string {
	return visualStyle.
		Width(v.width).
		Height(v.height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(v.stateIcon + " " + v.trackName)
}
