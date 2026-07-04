package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/yadneshx17/resonance/internal/common"
	"github.com/yadneshx17/resonance/internal/types"
)

type LibData struct {
	Entries []types.Entry
	Title   string
	Cursor  int
	Offset  int
	Active  bool
	Height  int
	Width   int
}

type Library struct {
	entries   []types.Entry
	title     string
	libCursor int
	libOffset int
	active    bool
	height    int
	width     int
}

var libPanel = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder())

var libActivePanel = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#7D56F4"))

func (l *Library) SetData(data LibData) {
	l.entries = data.Entries
	l.title = data.Title
	l.libCursor = data.Cursor
	l.libOffset = data.Offset
	l.active = data.Active
	l.height = data.Height
	l.width = data.Width
}

func (l Library) buildLibBlock() string {
	if l.width < 4 {
		return ""
	}
	var s string
	s += common.HeaderStyle.Render("Library: "+l.title) + "\n"

	var bar string
	for i := 0; i < l.width-2; i++ {
    	bar += "─"
	}
	s += bar + "\n"

	if len(l.entries) == 0 {
		s += " Empty\n"
		return s
	}
	contentWidth := l.width - 2 // 2 borders
	nameMax := contentWidth - 5 // prefix + "" + icon + ""
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
			name = string(runes[:nameMax-3]) + "…"
		}
		lines = append(lines, fmt.Sprintf("%s%s %s", prefix, icon, name))
	}
	s += strings.Join(lines, "\n")
	return s
}

func (l Library) View() string {
	content := l.buildLibBlock()
	style := libPanel
	if l.active {
		style = libActivePanel
	}
	return style.
		Width(l.width).
		Height(l.height).
		Render(content)
}
