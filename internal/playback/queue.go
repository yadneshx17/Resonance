package playback

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/gopxl/beep/mp3"
)

type Queue struct {
	tracks  []Track
	current int
}

func NewQueue() *Queue {
	return &Queue{}
}

func (q *Queue) Add(track Track) {
	q.tracks = append(q.tracks, track)
}

func (q *Queue) Next() (Track, bool) {
	if q.current+1 >= len(q.tracks) {
		return Track{}, false
	}
	q.current++
	return q.tracks[q.current], true
}

func (q *Queue) Prev() (Track, bool) {
	if q.current-1 < 0 {
		return Track{}, false
	}
	q.current--
	return q.tracks[q.current], true
}

func (q *Queue) Current() (Track, bool) {
	if len(q.tracks) == 0 {
		return Track{}, false
	}
	return q.tracks[q.current], true
}

func (q *Queue) SetCurrent(i int) {
	if i >= 0 && i < len(q.tracks) {
		q.current = i
	}
}

func (q *Queue) Remove(i int) {
	if i < 0 || i >= len(q.tracks) {
		return
	}
	q.tracks = append(q.tracks[:i], q.tracks[i+1:]...)
	if i < q.current {
		q.current--
	} else if q.current >= len(q.tracks) {
		q.current = max(0, len(q.tracks)-1)
	}
}

func (q *Queue) Len() int {
	return len(q.tracks)
}

func (q *Queue) Clear() {
	q.tracks = nil
	q.current = 0
}

func (q *Queue) List() []Track {
	result := make([]Track, len(q.tracks))
	copy(result, q.tracks)
	return result
}

func (q *Queue) CurrentIndex() int {
	return q.current
}

func (q *Queue) Shuffle() {
	rand.Shuffle(len(q.tracks), func(i, j int) {
		q.tracks[i], q.tracks[j] = q.tracks[j], q.tracks[i]
	})
}

func (q *Queue) ScanDir(dir string) ([]Track, error) {
	dirPath := filepath.Join("..", "..", dir)
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return []Track{}, err
	}

	var tracks []Track
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fpath := filepath.Join(dirPath, file.Name())
		f, err := os.Open(fpath)
		if err != nil {
			continue
		}

		_, _, err = mp3.Decode(f)
		if err != nil {
			f.Close()
			fmt.Printf("\nSkipping %s: %v", file.Name(), err)
			continue
		}
		f.Close()

		tracks = append(tracks, Track{Path: file.Name()})
	}
	return tracks, nil
}
