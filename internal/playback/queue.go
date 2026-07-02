package playback

import (
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopxl/beep/mp3"
	"github.com/yadneshx17/resonance/internal/types"
)

type Queue struct {
	tracks  []types.Track
	current int
}

func NewQueue() *Queue {
	return &Queue{}
}

func (q *Queue) Add(track types.Track) {
	q.tracks = append(q.tracks, track)
}

func (q *Queue) Next() (types.Track, bool) {
	if q.current+1 >= len(q.tracks) {
		return types.Track{}, false
	}
	q.current++
	return q.tracks[q.current], true
}

func (q *Queue) Prev() (types.Track, bool) {
	if q.current-1 < 0 {
		return types.Track{}, false
	}
	q.current--
	return q.tracks[q.current], true
}

func (q *Queue) Current() (types.Track, bool) {
	if len(q.tracks) == 0 {
		return types.Track{}, false
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

func (q *Queue) List() []types.Track {
	result := make([]types.Track, len(q.tracks))
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

// recursive directory transversal.
func (q *Queue) ScanDir(root string) ([]types.Track, error) {
	var tracks []types.Track
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			// skips and continue the scanning.
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".mp3") {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		_, _, err = mp3.Decode(f)
		if err != nil {
			f.Close()
			return nil
		}
		// defer in loops leaks file handles
		f.Close()
		tracks = append(tracks, types.Track{Path: path})
		return nil
	})
	return tracks, err
}