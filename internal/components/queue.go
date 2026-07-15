package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/yadneshx17/resonance/internal/common"
	"github.com/yadneshx17/resonance/internal/renderer"
	"github.com/yadneshx17/resonance/internal/types"
)

type QueueData struct {
	Tracks      []types.Track
	PlayingIdx  int
	Playing     bool
	Cursor      int
	Offset      int
	Total       int
	Active      bool
	Height      int
	Width       int
	SearchMode  bool
	SearchQuery string
}

type Queue struct {
	tracks      []types.Track
	playingIdx  int
	playing     bool
	cursor      int
	offset      int
	total       int
	active      bool
	height      int
	width       int
	searchMode  bool
	searchQuery string
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
	q.searchMode = data.SearchMode
	q.searchQuery = data.SearchQuery
}

func (q Queue) buildQueueBlock() string {
	if q.width < 4 {
		return ""
	}
	contentWidth := q.width - 4
	var lines []string
	if q.searchMode {
		lines = append(lines, common.SearchBar(q.searchQuery, contentWidth))
	}
	if len(q.tracks) == 0 {
		if q.searchMode {
			return strings.Join(lines, "\n")
		}
		style := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center)
		return style.Render("Nothing in Queue")
	}
	durWidth := 6
	seqWidth := 3
	prefixWidth := 3
	nameMax := contentWidth - seqWidth - 1 - prefixWidth - 1 - durWidth
	if nameMax < 1 {
		nameMax = 1
	}
	for i, t := range q.tracks {
		idx := q.offset + i
		name := t.Title
		if name == "" {
			name = t.Path
			if slash := strings.LastIndexByte(t.Path, '/'); slash >= 0 {
				name = t.Path[slash+1:]
			}
		} else if t.Artist != "" {
			name += " - " + t.Artist
		}
		runes := []rune(name)
		if len(runes) > nameMax {
			name = string(runes[:nameMax-1]) + "…"
		}
		seq := fmt.Sprintf("%*d", seqWidth-1, idx+1)
		dur := ""
		if t.Duration > 0 {
			dur = common.FmtDuration(t.Duration)
		}
		durFmt := fmt.Sprintf("%*s", durWidth, dur)
		prefix := "   "
		if idx == q.playingIdx && q.playing {
			prefix = common.Play + "  "
		} else if q.active && q.cursor == idx {
			prefix = common.Cursor + " "
		}
		padded := fmt.Sprintf("%-*s", nameMax, name)
		line := fmt.Sprintf("%s %s%s %s", seq, prefix, padded, durFmt)
		if idx == q.playingIdx && q.playing {
			line = common.PlayingTrackStyle.Render(line)
		} else if q.active && q.cursor == idx {
			line = common.CursorStyle.Render(line)
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
