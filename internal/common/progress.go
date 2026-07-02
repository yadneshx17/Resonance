package common

import (
	"fmt"
	"time"
)

func FmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", m, s)
}

func ProgressBar(pos, dur time.Duration, width int) string {
	if dur == 0 {
		return ""
	}
	filled := int(float64(pos) / float64(dur) * float64(width))
	if filled > width {
		filled = width
	}
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}
