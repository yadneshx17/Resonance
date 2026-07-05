package components

import (
	"fmt"
	"strings"

	"github.com/yadneshx17/resonance/internal/common"
	"github.com/yadneshx17/resonance/internal/renderer"
	"github.com/yadneshx17/resonance/internal/types"
)

type LibData struct {
	Entries []types.Entry
	Title   string
	Cursor  int
	Offset  int
	Total   int
	Active  bool
	Height  int
	Width   int
}

type Library struct {
	entries   []types.Entry
	title     string
	libCursor int
	libOffset int
	total     int
	active    bool
	height    int
	width     int
}

func (l *Library) SetData(data LibData) {
	l.entries = data.Entries
	l.title = data.Title
	l.libCursor = data.Cursor
	l.libOffset = data.Offset
	l.total = data.Total
	l.active = data.Active
	l.height = data.Height
	l.width = data.Width
}

func (l Library) buildLibBlock() string {
	if l.width < 4 {
		return ""
	}
	if len(l.entries) == 0 {
		return " Empty"
	}
	contentWidth := l.width - 4
	nameMax := contentWidth - 5
	if nameMax < 1 {
		nameMax = 1
	}
	var lines []string
	for i, e := range l.entries {
		idx := l.libOffset + i
		prefix := "  "
		if l.active && l.libCursor == idx {
			prefix = common.CursorStyle.Render(common.Cursor)
		}
		icon := common.Music
		if e.IsDir {
			icon = common.Directory
		}
		name := e.Name
		runes := []rune(name)
		if len(runes) > nameMax {
			name = string(runes[:nameMax-1]) + "…"
		}
		lines = append(lines, fmt.Sprintf("%s%s %s", prefix, icon, name))
	}
	return strings.Join(lines, "\n")
}

func (l Library) View() string {
	content := l.buildLibBlock()
	pos := l.libCursor + 1
	if l.total == 0 {
		pos = 0
	}
	info := []string{fmt.Sprintf("%d of %d", pos, l.total)}
	return renderer.Render(content, renderer.Config{
		Width:     l.width,
		Height:    l.height,
		Title:     "Library - " + l.title,
		InfoItems: info,
		Active:    l.active,
	})
}
