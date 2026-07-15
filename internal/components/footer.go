package components

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/yadneshx17/resonance/internal/common"
)

type FooterData struct {
	TrackName string
	StateIcon string
	Position  time.Duration
	Duration  time.Duration
	VolLevel  float64
	Muted     bool
	Height    int
	Width     int
	CoverArt  []byte
}

type Footer struct {
	trackName string
	stateIcon string
	position  time.Duration
	duration  time.Duration
	volLevel  float64
	muted     bool
	height    int
	width     int
	coverArt  []byte
}

var footerStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder())

func (f *Footer) SetData(data FooterData) {
	f.trackName = data.TrackName
	f.stateIcon = data.StateIcon
	f.position = data.Position
	f.duration = data.Duration
	f.volLevel = data.VolLevel
	f.muted = data.Muted
	f.height = data.Height
	f.width = data.Width
	f.coverArt = data.CoverArt
}

func (f *Footer) View() string {
	content := f.buildFooterBlock()
	return footerStyle.
		Width(f.width).
		Height(f.height).
		Render(content)
}

func (f Footer) buildFooterBlock() string {
	contentWidth := f.width - 2
	if contentWidth < 4 {
		return ""
	}

	innerLines := f.height - 2
	if innerLines < 1 {
		innerLines = 1
	}
	coverW := 16
	coverH := innerLines
	if coverH > 8 {
		coverH = 8
	}
	coverLines := make([]string, coverH)
	if len(f.coverArt) > 0 {
		coverStr := common.RenderCover(f.coverArt, coverW, coverH)
		if coverStr != "" {
			parts := strings.Split(coverStr, "\n")
			for i := 0; i < coverH && i < len(parts); i++ {
				coverLines[i] = parts[i]
			}
		}
	}
	textWidth := contentWidth - coverW - 2
	if textWidth < 10 {
		textWidth = contentWidth
		coverW = 0
	}

	info := f.stateIcon + " " + f.trackName
	runes := []rune(info)
	if len(runes) > textWidth {
		info = string(runes[:textWidth-1]) + "…"
	}
	info = fmt.Sprintf("%-*s", textWidth, info)

	timeStr := ""
	if f.duration > 0 {
		timeStr = fmt.Sprintf("%s / %s", common.FmtDuration(f.position), common.FmtDuration(f.duration))
	}
	barWidth := (textWidth - len(timeStr) - 1) / 2
	if barWidth < 0 {
		barWidth = 0
	}
	bar := common.ProgressBar(f.position, f.duration, barWidth)
	barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#89DCEB"))
	bar = barStyle.Render(bar)
	progressLine := bar
	if timeStr != "" {
		if bar != "" {
			progressLine += " "
		}
		progressLine += timeStr
	}
	progressLine = fmt.Sprintf("%-*s", textWidth, progressLine)

	volPct := int((f.volLevel + 3) / 6 * 100)
	if volPct < 0 {
		volPct = 0
	} else if volPct > 100 {
		volPct = 100
	}
	volStr := ""
	if f.muted {
		mutedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F38BA8"))
		volStr = mutedStyle.Render("[MUTED]")
	} else {
		volIcon := common.VolumeLow
		if volPct > 33 {
			volIcon = common.VolumeMedium + " "
		}
		if volPct > 66 {
			volIcon = common.VolumeHigh + " "
		}
		volStr = fmt.Sprintf("%svol: %d%%", volIcon, volPct)
	}
	volStr = fmt.Sprintf("%-*s", textWidth, volStr)

	textRows := []string{info, progressLine, volStr}
	for len(textRows) < innerLines {
		textRows = append(textRows, "")
	}

	lines := make([]string, innerLines)
	for i := 0; i < innerLines; i++ {
		text := ""
		if i < len(textRows) {
			text = textRows[i]
		}
		sep := "\x1b[38;2;69;71;90m│\x1b[39m  "
		if coverW > 0 && i < len(coverLines) && coverLines[i] != "" {
			lines[i] = coverLines[i] + "  " + sep + text
		} else if coverW > 0 && i < len(coverLines) {
			lines[i] = strings.Repeat(" ", coverW) + "  " + sep + text
		} else {
			lines[i] = text
		}
	}
	return strings.Join(lines, "\n")
}
