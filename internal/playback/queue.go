package playback

import (
	"fmt"
	"math/rand"
)

type Queue struct {
	tracks    []Track
	current   int
	playOrder []int
}

type QueueService interface {
	Add(track Track)
	Next() (Track, bool)
	Prev() (Track, bool)
	Current() (Track, bool)
	Len() int
	Clear()
	List() []Track
	Shuffle()
}

// Constructor
func NewQueue() *Queue {
	return &Queue{}
}

func (q *Queue) Add(track Track) {
	q.tracks = append(q.tracks, track)
	q.playOrder = append(q.playOrder, len(q.tracks)-1)
}

func (q *Queue) Next() (Track, bool) {
	if q.current+1 >= len(q.tracks) {
		return Track{}, false
	}
	q.current++
	return q.tracks[q.playOrder[q.current]], true
}

func (q *Queue) Prev() (Track, bool) {
	if q.current-1 < 0 {
		return Track{}, false
	}
	q.current--
	return q.tracks[q.playOrder[q.current]], true
}

func (q *Queue) Current() (Track, bool) {
	if len(q.tracks) == 0 {
		return Track{}, false
	}
	return q.tracks[q.playOrder[q.current]], true
}

func (q *Queue) Len() int {
	return len(q.tracks)
}

func (q *Queue) Clear() {
	q.tracks = nil
	q.playOrder = nil
	q.current = 0
}

func (q *Queue) List() []Track {
	result := make([]Track, len(q.tracks))
	for i, idx := range q.playOrder {
		result[i] = q.tracks[idx]
	}
	return result
}

func (q *Queue) PrintList() {
	tracks := q.List()
	fmt.Println("\n>> Tracks")
	for idx, tracks := range tracks {
		track := tracks.Path
		fmt.Printf("\n   %v  %v", idx, track)
	}
}

func (q *Queue) Shuffle() {
	for i := range q.playOrder {
		j := i + int(rand.Intn(len(q.playOrder)-i))
		q.playOrder[i], q.playOrder[j] = q.playOrder[j], q.playOrder[i]
	}
}
