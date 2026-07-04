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

	info := "" + f.stateIcon + " " + f.trackName + " "
	runes := []rune(info)
	if len(runes) > contentWidth {
		info = string(runes[:contentWidth-1]) + "…"
	}

	timeStr := ""
	if f.duration > 0 {
		timeStr = fmt.Sprintf("%s / %s", common.FmtDuration(f.position), common.FmtDuration(f.duration))
	}
	barWidth := contentWidth - len(timeStr) - 1
	if barWidth < 0 {
		barWidth = 0
	}
	bar := common.ProgressBar(f.position, f.duration, barWidth)
	progressLine := bar
	if timeStr != "" {
		if bar != "" {
			progressLine += " "
		}
		progressLine += timeStr
	}

	volPct := int((f.volLevel + 3) / 6 * 100)
	if volPct < 0 {
		volPct = 0
	} else if volPct > 100 {
		volPct = 100
	}
	volStr := ""
	if f.muted {
		volStr = "[MUTE]"
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
	// controlsLine := fmt.Sprintf("%s  %s  %s  │  %s", common.PrevTrack, f.stateIcon, common.NextTrack, volStr)
	controlsLine := fmt.Sprintf("%s", volStr)

	return strings.Join([]string{info, progressLine, controlsLine}, "\n")
}
