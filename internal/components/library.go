package components

import (
	"fmt"
	"strings"

	"github.com/yadneshx17/resonance/internal/common"
	"github.com/yadneshx17/resonance/internal/renderer"
	"github.com/yadneshx17/resonance/internal/types"
)

type LibData struct {
	Entries     []types.Entry
	Title       string
	Cursor      int
	Offset      int
	Total       int
	Active      bool
	Height      int
	Width       int
	SearchMode  bool
	SearchQuery string
	SelectedPath string
}

type Library struct {
	entries      []types.Entry
	title        string
	libCursor    int
	libOffset    int
	total        int
	active       bool
	height       int
	width        int
	searchMode   bool
	searchQuery  string
	selectedPath string
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
	l.searchMode = data.SearchMode
	l.searchQuery = data.SearchQuery
	l.selectedPath = data.SelectedPath
}

func (l Library) buildLibBlock() string {
	if l.width < 4 {
		return ""
	}
	contentWidth := l.width - 4
	var lines []string
	if l.searchMode {
		lines = append(lines, common.SearchBar(l.searchQuery, contentWidth))
	}
	dirs := make([]types.Entry, 0, len(l.entries))
	for _, e := range l.entries {
		if e.IsDir {
			dirs = append(dirs, e)
		}
	}
	if len(dirs) == 0 {
		if l.searchMode {
			return strings.Join(lines, "\n")
		}
		return " Empty"
	}
	nameMax := contentWidth - 5
	if nameMax < 1 {
		nameMax = 1
	}
	for i, e := range dirs {
		idx := l.libOffset + i
		name := e.Name
		runes := []rune(name)
		if len(runes) > nameMax {
			name = string(runes[:nameMax-1]) + "…"
		}
		selected := l.selectedPath != "" && e.Path == l.selectedPath
		line := common.Directory + " " + name
		if selected {
			line = common.PlayingTrackStyle.Render("▶ " + name)
		} else if l.active && l.libCursor == idx {
			line = common.CursorStyle.Render(common.Cursor + " " + line)
		} else {
			line = "  " + line
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (l Library) View() string {
	content := l.buildLibBlock()
	dirs := 0
	for _, e := range l.entries {
		if e.IsDir {
			dirs++
		}
	}
	pos := l.libCursor + 1
	if dirs == 0 {
		pos = 0
	}
	info := []string{fmt.Sprintf("%d of %d", pos, dirs)}
	return renderer.Render(content, renderer.Config{
		Width:     l.width,
		Height:    l.height,
		Title:     "Library - " + l.title,
		InfoItems: info,
		Active:    l.active,
	})
}
