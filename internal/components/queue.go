package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/yadneshx17/resonance/internal/common"
	"github.com/yadneshx17/resonance/internal/types"
)

type QueueData struct {
	Tracks     []types.Track
	PlayingIdx int
	Playing    bool
	Cursor     int
	Offset     int
	Active     bool
	Height     int
	Width      int
}

type Queue struct {
	tracks     []types.Track
	playingIdx int
	playing    bool
	cursor     int
	offset     int
	active     bool
	height     int
	width      int
}

var queuePanel = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder())

var queueActivePanel = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#7D56F4"))

func (q *Queue) SetData(data QueueData) {
	q.tracks = data.Tracks
	q.playingIdx = data.PlayingIdx
	q.playing = data.Playing
	q.cursor = data.Cursor
	q.offset = data.Offset
	q.active = data.Active
	q.height = data.Height
	q.width = data.Width
}

func (q Queue) buildQueueBlock() string {
	if q.width < 4 {
		return ""
	}
	var s string
	s += common.HeaderStyle.Render("Queue") + "\n"
	var bar string
	for i := 0; i < q.width-2; i++ {
    	bar += "─"
	}
	s += bar + "\n"
	if len(q.tracks) == 0 {
		s += " Empty\n"
		return s
	}
	contentWidth := q.width - 2
	nameMax := contentWidth - 2
	if nameMax < 1 {
		nameMax = 1
	}
	var lines []string
	for i, t := range q.tracks {
		idx := q.offset + i
		name := t.Path
		if slash := strings.LastIndexByte(t.Path, '/'); slash >= 0 {
			name = t.Path[slash+1:]
		}
		runes := []rune(name)
		if len(runes) > nameMax {
			name = string(runes[:nameMax-3]) + "…"
		}
		line := fmt.Sprintf("  %s", name)
		if idx == q.playingIdx && q.playing {
			line = common.PlayingTrackStyle.Render(fmt.Sprintf("%s %s", common.Play, name))
		} else if q.active && q.cursor == idx {
			line = common.CursorStyle.Render(common.Cursor + "" + name)
		}
		lines = append(lines, line)
	}
	s += strings.Join(lines, "\n")
	return s
}

func (q Queue) View() string {
	content := q.buildQueueBlock()
	style := queuePanel
	if q.active {
		style = queueActivePanel
	}
	return style.
		Width(q.width).
		Height(q.height).
		Render(content)
}
