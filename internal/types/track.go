package types

import "time"

type Track struct {
	ID          string
	Title       string
	Artist      string
	Album       string
	CoverArt    []byte // raw JPEG/PNG bytes
	CoverArtURL string // remote album-art URL (Spotify)
	Source      Source // enum
	Path        string // local filesystem Path
	Duration    time.Duration
}

type Source int

const (
	Local Source = iota
	Spotify
)
