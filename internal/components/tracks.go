package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/yadneshx17/resonance/internal/common"
	"github.com/yadneshx17/resonance/internal/renderer"
	"github.com/yadneshx17/resonance/internal/types"
)

type TracksData struct {
	Entries     []types.Entry
	PlayingPath string
	Cursor      int
	Offset      int
	Total       int
	Active      bool
	Height      int
	Width       int
	SearchMode  bool
	SearchQuery string
	SubdirCount int
}

type Tracks struct {
	entries     []types.Entry
	playingPath string
	cursor      int
	offset      int
	total       int
	active      bool
	height      int
	width       int
	searchMode  bool
	searchQuery string
	subdirCount int
}

func (t *Tracks) SetData(data TracksData) {
	t.entries = data.Entries
	t.playingPath = data.PlayingPath
	t.cursor = data.Cursor
	t.offset = data.Offset
	t.total = data.Total
	t.active = data.Active
	t.height = data.Height
	t.width = data.Width
	t.searchMode = data.SearchMode
	t.searchQuery = data.SearchQuery
	t.subdirCount = data.SubdirCount
}

func (t Tracks) buildBlock() string {
	if t.width < 4 {
		return ""
	}
	contentWidth := t.width - 4
	var lines []string
	if t.searchMode {
		lines = append(lines, common.SearchBar(t.searchQuery, contentWidth))
	}
	if len(t.entries) == 0 {
		if t.searchMode {
			return strings.Join(lines, "\n")
		}
		style := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center)
		if t.subdirCount > 0 {
			msg := fmt.Sprintf("Contains %d subdirector", t.subdirCount)
			if t.subdirCount == 1 {
				msg += "y"
			} else {
				msg += "ies"
			}
			return style.Render(msg)
		}
		return style.Render("No tracks in this directory")
	}
	seqWidth := 3
	prefixWidth := 3
	durWidth := 6
	remaining := contentWidth - seqWidth - 1 - prefixWidth - 3 - durWidth
	titleWidth := remaining * 34 / 100
	artistWidth := remaining * 33 / 100
	albumWidth := remaining - titleWidth - artistWidth
	if titleWidth < 3 {
		titleWidth = 3
	}
	if artistWidth < 3 {
		artistWidth = 3
	}
	if albumWidth < 3 {
		albumWidth = 3
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#6C7086"))
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#45475A"))

	header := fmt.Sprintf("%*s %-*s%-*s %-*s %-*s %*s",
		seqWidth, "#",
		prefixWidth, "",
		titleWidth, "Title",
		artistWidth, "Artist",
		albumWidth, "Album",
		durWidth, "Dur",
	)
	header = headerStyle.Render(header)

	sep := fmt.Sprintf("%s %s%s %s %s %s",
		strings.Repeat("─", seqWidth),
		strings.Repeat("─", prefixWidth),
		strings.Repeat("─", titleWidth),
		strings.Repeat("─", artistWidth),
		strings.Repeat("─", albumWidth),
		strings.Repeat("─", durWidth),
	)
	sep = sepStyle.Render(sep)

	lines = append(lines, header, sep)

	for i, e := range t.entries {
		idx := t.offset + i
		title := e.Title
		if title == "" {
			title = e.Name
		}
		artist := e.Artist
		album := e.Album
		runesT := []rune(title)
		if len(runesT) > titleWidth {
			title = string(runesT[:titleWidth-1]) + "…"
		}
		runesA := []rune(artist)
		if len(runesA) > artistWidth {
			artist = string(runesA[:artistWidth-1]) + "…"
		}
		runesAl := []rune(album)
		if len(runesAl) > albumWidth {
			album = string(runesAl[:albumWidth-1]) + "…"
		}
		seq := fmt.Sprintf("%*d", seqWidth, idx+1)
		dur := ""
		if e.Duration > 0 {
			dur = common.FmtDuration(e.Duration)
		}
		durFmt := fmt.Sprintf("%*s", durWidth, dur)
		playing := t.playingPath != "" && e.Path == t.playingPath
		prefix := "   "
		if playing {
			prefix = common.Play + "  "
		} else if t.active && t.cursor == idx {
			prefix = common.Cursor
		}
		line := fmt.Sprintf("%s %s%-*s %-*s %-*s %s", seq, prefix, titleWidth, title, artistWidth, artist, albumWidth, album, durFmt)
		if playing {
			line = common.PlayingTrackStyle.Render(line)
		} else if t.active && t.cursor == idx {
			line = common.CursorStyle.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (t Tracks) View() string {
	content := t.buildBlock()
	pos := t.cursor + 1
	if t.total == 0 {
		pos = 0
	}
	info := []string{fmt.Sprintf("%d of %d", pos, t.total)}
	return renderer.Render(content, renderer.Config{
		Width:     t.width,
		Height:    t.height,
		Title:     "Content",
		InfoItems: info,
		Active:    t.active,
	})
}
