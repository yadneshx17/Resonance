package playback

import "time"

type Track struct {
	ID       string
	Title    string
	Path     string
	Duration time.Duration
}
