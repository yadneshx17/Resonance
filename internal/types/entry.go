package types

import "time"

type Entry struct {
	Name     string
	Path     string
	IsDir    bool
	Title    string
	Artist   string
	Album    string
	Duration time.Duration
	CoverArt []byte
}
