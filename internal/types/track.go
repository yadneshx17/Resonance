package types

import "time"

type Track struct {
	ID       string
	Title    string
	Artist   string
	Album    string
	CoverArt []byte // raw JPEG/PNG bytes
	Source   Source // enum
	Path     string // local filesystem Path
	Duration time.Duration
}

type Source int

const (
	SourceLocal Source = iota
	SourceSpotify
	SourceJam
)
