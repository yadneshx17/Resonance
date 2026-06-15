package playback

type PlaybackState int

const (
	Stopped PlaybackState = iota
	Playing
	Paused
)
