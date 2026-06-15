package playback

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
)

// Stores the loaded audio and current state
type Player struct {
	streamer beep.StreamSeekCloser // decoded audio + ( Close, Seek, Position )
	format   beep.Format           // sample rate info
	ctrl     *beep.Ctrl            // enables pause/resume
	state    PlaybackState
	track    *Track
	done     chan struct{}
}

type AudioEngine interface {
	Load(track Track) error

	Wait()
	Play() error
	Pause() error
	Stop() error
	Resume() error

	Seek(d time.Duration) error // streamer.Seek(d)
	Position() time.Duration    // streamer.Position

	State() PlaybackState // Returns stored State
}

// Constructor
func NewPlayer() *Player {
	return &Player{
		state: Stopped,
	}
}

func (p *Player) Load(track Track) error {
	if p.streamer != nil {
		p.streamer.Close()
	}

	path := filepath.Join("..", "..", "Music", "shiv2.mp3")
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		f.Close()
		return err
	}
	p.streamer = streamer
	p.format = format
	p.track = &track
	p.state = Stopped
	fmt.Printf("File %v loaded", track.Path)
	return nil
}

func (p *Player) Wait() {
	if p.done != nil {
		<-p.done
	}
}

func (p *Player) Play() error {
	if p.streamer == nil {
		return fmt.Errorf("no track loaded")
	}

	speaker.Init(p.format.SampleRate, p.format.SampleRate.N(time.Second/10))

	p.ctrl = &beep.Ctrl{Streamer: p.streamer, Paused: false}
	p.done = make(chan struct{})

	speaker.Play(beep.Seq(p.ctrl, beep.Callback(func() { close(p.done) })))

	p.state = Playing

	fmt.Println("\nPlaying...")
	return nil
}

func (p *Player) Pause() error {
	if p.ctrl == nil {
		return fmt.Errorf("nothing playing")
	}
	p.ctrl.Paused = true
	p.state = Paused
	return nil
}

func (p *Player) Resume() error {
	if p.ctrl == nil {
		return fmt.Errorf("nothing playing")
	}
	p.ctrl.Paused = false
	p.state = Playing
	return nil
}

func (p *Player) Stop() error {
	speaker.Clear() // removes all currently playing Streamers from speaker
	p.state = Stopped
	if p.streamer != nil {
		p.streamer.Seek(0) // reset position
	}
	close(p.done)
	return nil
}

func (p *Player) Seek(position time.Duration) error {
	if p.streamer == nil {
		return fmt.Errorf("no track loaded")
	}
	posi := p.format.SampleRate.N(position)
	return p.streamer.Seek(posi)
}

func (p *Player) Position() time.Duration {
	if p.streamer == nil {
		return 0
	}
	return time.Duration(p.streamer.Position())
}

func (p *Player) State() PlaybackState {
	return p.state
}
