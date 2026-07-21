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
	Title     string
	Album     string
	Artist    string
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
	title     string
	album     string
	artist    string
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
	f.title = data.Title
	f.album = data.Album
	f.artist = data.Artist
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
		Bold(true).
		Italic(true).
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

	// Cover art
	coverW := 32
	coverH := innerLines
	if coverH > 9 {
		coverH = 9
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

	// Text area — "  │  " separator = 5 chars
	textWidth := contentWidth - coverW - 5
	if textWidth < 10 {
		textWidth = contentWidth
		coverW = 0
	}

	// Styles
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	dimBold := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Bold(true)
	bold := lipgloss.NewStyle().Bold(true)
	italic := lipgloss.NewStyle().Italic(true)
	barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#89DCEB"))
	mutedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F38BA8"))
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#45475A"))

	// Volume
	volPct := int((f.volLevel + 3) / 6 * 100)
	volPct = max(0, min(100, volPct))

	// ── Alignment anchors ──
	// Progress bar starts where title starts, ends where artist/album end.
	// Artist and album share a right edge with 4-col padding from border.
	rightPad := 4

	// Row 1: Time + Title + Artist
	timeText := fmt.Sprintf("+ Time %s/%s", common.FmtDuration(f.position), common.FmtDuration(f.duration))
	titleText := f.title
	artistText := "by " + f.artist

	timeW := len([]rune(timeText))
	artistW := len([]rune(artistText))

	gap1 := 4
	barStart := timeW + gap1
	barEnd := textWidth - rightPad
	barWidth := barEnd - barStart
	if barWidth < 4 {
		barWidth = 4
		barEnd = barStart + barWidth
	}

	titleMax := barWidth - artistW - 1
	if titleMax < 1 {
		titleMax = 1
	}
	if len([]rune(titleText)) > titleMax {
		titleText = string([]rune(titleText)[:titleMax-1]) + "…"
	}
	titleW := len([]rune(titleText))

	gap2 := barWidth - titleW - artistW
	if gap2 < 1 {
		gap2 = 1
	}

	row1 := dimBold.Render(timeText) +
		strings.Repeat(" ", gap1) +
		bold.Render(titleText) +
		strings.Repeat(" ", gap2) +
		dim.Bold(true).Italic(true).Render(artistText)

	// Row 2: Volume (left) + Album (right-aligned to barEnd)
	var volText string
	if f.muted {
		volText = "- Vol [MUTED]"
	} else {
		volText = fmt.Sprintf("- Vol %d%%", volPct)
	}

	volW := len([]rune(volText))

	albumText := f.album
	albumAvail := barEnd - volW - 1
	if albumAvail < 1 {
		albumAvail = 1
	}
	if len([]rune(albumText)) > albumAvail {
		albumText = string([]rune(albumText)[:albumAvail-1]) + "…"
	}
	albumW := len([]rune(albumText))

	gap3 := barEnd - volW - albumW
	if gap3 < 1 {
		gap3 = 1
	}

	var volStyled string
	if f.muted {
		volStyled = mutedStyle.Render(volText)
	} else {
		volStyled = dim.Render(volText)
	}

	row2 := volStyled +
		strings.Repeat(" ", gap3) +
		italic.Render(albumText)

	// Row 3: Controls (below Vol) + Progress bar
	controlsStyled := "  " + dim.Render(common.PrevTrack) + "  " + f.stateIcon + "  " + dim.Render(common.NextTrack)
	controlsW := 9
	pbarStart := barStart
	pbarWidth := barEnd - pbarStart
	if pbarWidth < 1 {
		pbarWidth = 1
	}
	bar := common.ProgressBar(f.position, f.duration, pbarWidth)
	row3 := controlsStyled + strings.Repeat(" ", pbarStart-controlsW) + barStyle.Render(bar)

	// Assemble text rows
	textRows := []string{row1, row2, row3}
	for len(textRows) < innerLines {
		textRows = append(textRows, "")
	}

	// Combine cover art and text
	lines := make([]string, innerLines)
	sep := sepStyle.Render("│")
	for i := 0; i < innerLines; i++ {
		text := ""
		if i < len(textRows) {
			text = textRows[i]
		}
		if coverW > 0 && i < coverH {
			if coverLines[i] != "" {
				lines[i] = coverLines[i] + "  " + sep + "  " + text
			} else {
				lines[i] = strings.Repeat(" ", coverW) + "  " + sep + "  " + text
			}
		} else if coverW > 0 {
			lines[i] = strings.Repeat(" ", coverW) + "  " + sep + "  " + text
		} else {
			lines[i] = text
		}
	}
	return strings.Join(lines, "\n")
}
