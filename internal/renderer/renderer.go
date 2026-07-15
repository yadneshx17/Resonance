package renderer

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/mattn/go-runewidth"
)

type Config struct {
	Width     int
	Height    int
	Title     string
	InfoItems []string
	Active    bool
}

// truncateCells truncates s to at most maxCells visible cells preserving ANSI.
func truncateCells(s string, maxCells int) string {
	if maxCells < 1 {
		return ""
	}
	cells := 0
	ri := 0
	for ri < len(s) {
		r, size := utf8.DecodeRuneInString(s[ri:])
		if r == '\x1b' {
			ri += size
			for ri < len(s) {
				b := s[ri]
				ri++
				if b == 'm' {
					break
				}
			}
			continue
		}
		w := runewidth.RuneWidth(r)
		if cells+w > maxCells {
			return s[:ri] + "…"
		}
		cells += w
		ri += size
	}
	return s
}

func Render(content string, cfg Config) string {
	w := cfg.Width
	if w < 8 || cfg.Height < 3 {
		return content
	}

	r, g, b := 69, 71, 90
	if cfg.Active {
		r, g, b = 166, 227, 161
	}
	bc := fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
	nc := "\x1b[39m"

	cw := w - 4

	titleW := lipgloss.Width(cfg.Title)
	topDash := w - 6 - titleW
	if topDash < 0 {
		topDash = 0
	}
	top := bc + "╭── " + cfg.Title + " " + strings.Repeat("─", topDash) + "╮" + nc

	lines := strings.Split(content, "\n")
	var body []string
	for _, line := range lines {
		lw := lipgloss.Width(line)
		if lw > cw {
			line = truncateCells(line, cw)
		} else if lw < cw {
			line += strings.Repeat(" ", cw-lw)
		}
		body = append(body, bc+"│ "+nc+line+bc+" │"+nc)
	}

	need := cfg.Height - 2 - len(body)
	for i := 0; i < need; i++ {
		body = append(body, bc+"│"+nc+strings.Repeat(" ", cw+2)+bc+"│"+nc)
	}

	bottom := bc
	if len(cfg.InfoItems) > 0 {
		info := strings.Join(cfg.InfoItems, " │ ")
		iw := lipgloss.Width(info)
		bd := w - 4 - iw
		if bd < 0 {
			info = truncateCells(info, w-8)
			iw = lipgloss.Width(info)
			bd = w - 6 - iw
			if bd < 0 {
				bd = 0
			}
		}
		bottom += "╰" + strings.Repeat("─", bd) + " " + info + " " + "╯"
	} else {
		bottom += "╰" + strings.Repeat("─", w-2) + "╯"
	}
	bottom += nc

	all := append([]string{top}, body...)
	all = append(all, bottom)
	return strings.Join(all, "\n")
}
