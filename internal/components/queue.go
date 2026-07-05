package components

import (
	"fmt"
	"strings"

	"github.com/yadneshx17/resonance/internal/common"
	"github.com/yadneshx17/resonance/internal/renderer"
	"github.com/yadneshx17/resonance/internal/types"
)

type QueueData struct {
	Tracks     []types.Track
	PlayingIdx int
	Playing    bool
	Cursor     int
	Offset     int
	Total      int
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
	total      int
	active     bool
	height     int
	width      int
}

func (q *Queue) SetData(data QueueData) {
	q.tracks = data.Tracks
	q.playingIdx = data.PlayingIdx
	q.playing = data.Playing
	q.cursor = data.Cursor
	q.offset = data.Offset
	q.total = data.Total
	q.active = data.Active
	q.height = data.Height
	q.width = data.Width
}

func (q Queue) buildQueueBlock() string {
	if q.width < 4 {
		return ""
	}
	if len(q.tracks) == 0 {
		return " Empty"
	}
	contentWidth := q.width - 4
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
			name = string(runes[:nameMax-1]) + "…"
		}
		line := fmt.Sprintf("  %s", name)
		if idx == q.playingIdx && q.playing {
			line = common.PlayingTrackStyle.Render(fmt.Sprintf("%s %s", common.Play, name))
		} else if q.active && q.cursor == idx {
			line = common.CursorStyle.Render(common.Cursor + " " + name)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (q Queue) View() string {
	content := q.buildQueueBlock()
	pos := q.cursor + 1
	if q.total == 0 {
		pos = 0
	}
	info := []string{fmt.Sprintf("%d of %d", pos, q.total)}
	return renderer.Render(content, renderer.Config{
		Width:     q.width,
		Height:    q.height,
		Title:     "Queue",
		InfoItems: info,
		Active:    q.active,
	})
}
