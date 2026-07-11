package renderer

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

type Config struct {
	Width     int
	Height    int
	Title     string
	InfoItems []string
	Active    bool
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
			runes := []rune(line)
			line = string(runes[:cw-1]) + "…"
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
			runes := []rune(info)
			info = string(runes[:w-8]) + "…"
			iw = lipgloss.Width(info)
			bd = w - 6 - iw
			if bd < 0 {
				bd = 0
			}
		}
		// bottom += "╰" + strings.Repeat("─", bd) + "┤" + info + "├
		bottom += "╰" + strings.Repeat("─", bd) + " " + info + " " + "╯"
	} else {
		bottom += "╰" + strings.Repeat("─", w-2) + "╯"
	}
	bottom += nc

	all := append([]string{top}, body...)
	all = append(all, bottom)
	return strings.Join(all, "\n")
}
