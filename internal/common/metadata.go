package common

import (
	"os"
	"time"

	"github.com/dhowden/tag"
	"github.com/gopxl/beep/mp3"
)

type Metadata struct {
	Title    string
	Artist   string
	Album    string
	Lyrics   string
	Duration time.Duration
	CoverArt []byte
}

func ReadMetadata(path string) (Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return Metadata{}, nil
	}
	streamer, format, err := mp3.Decode(f)
	if err != nil {
		f.Close()
		return Metadata{}, nil
	}
	samples := streamer.Len()
	dur := time.Duration(float64(samples) / float64(format.SampleRate) * float64(time.Second))
	streamer.Close()

	m, err := os.Open(path)
	if err != nil {
		return Metadata{Duration: dur}, nil
	}
	defer m.Close()
	meta, err := tag.ReadFrom(m)
	if err != nil {
		return Metadata{Duration: dur}, nil
	}
	t := Metadata{
		Title:    meta.Title(),
		Artist:   meta.Artist(),
		Album:    meta.Album(),
		Duration: dur,
	}
	if cover := meta.Picture(); cover != nil {
		t.CoverArt = cover.Data
	}
	return t, nil
}
