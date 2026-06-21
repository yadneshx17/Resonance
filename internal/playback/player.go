package playback

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
)

// Stores the loaded audio and current state
type Player struct {
	streamer  beep.StreamSeekCloser // decoded audio + ( Close, Seek, Position )
	format    beep.Format           // sample rate info
	ctrl      *beep.Ctrl            // enables pause/resume
	volume    *effects.Volume       // volume control
	state     PlaybackState
	track     *Track
	done      chan struct{}
	closeOnce sync.Once            // fixes: panic closing closed channel
	volLevel  float64              // current volume level (-3 to 3)
	muted     bool
	prevVol   float64              // volume before mute
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

func (p *Player) CurrentTrack() Track {
	if p.track == nil {
		return Track{}
	}
	return *p.track
}

func (p *Player) SetVolume(delta float64) {
	p.volLevel += delta
	if p.volLevel < -3 {
		p.volLevel = -3
	}
	if p.volLevel > 3 {
		p.volLevel = 3
	}
	if p.volume != nil {
		p.volume.Volume = p.volLevel
	}
}

func (p *Player) Volume() float64 {
	return p.volLevel
}

func (p *Player) Mute() {
	p.muted = true
	if p.volume != nil {
		p.volume.Silent = true
	}
}

func (p *Player) Unmute() {
	p.muted = false
	if p.volume != nil {
		p.volume.Silent = false
	}
}

func (p *Player) IsMuted() bool {
	return p.muted
}

func (p *Player) Load(track Track) error {
	if p.streamer != nil {
		p.streamer.Close()
	}

	f, err := os.Open(track.Path)
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
	return nil
}

func (p *Player) Wait() {
	if p.done != nil {
		<-p.done // unblock
	}
}

func (p *Player) Play() error {
	if p.streamer == nil {
		return fmt.Errorf("no track loaded")
	}

	// fix: This should be decoupled and Init at application startup and not for every song.
	speaker.Init(p.format.SampleRate, p.format.SampleRate.N(time.Second/10))

	p.ctrl = &beep.Ctrl{Streamer: p.streamer, Paused: false}
	p.volume = &effects.Volume{Streamer: p.ctrl, Base: 2, Volume: p.volLevel, Silent: p.muted}

	p.closeOnce = sync.Once{}
	p.done = make(chan struct{})

	done := p.done
	speaker.Play(beep.Seq(p.volume, beep.Callback(func() {
		close(done)
	})))

	p.state = Playing
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

	p.closeOnce.Do(func() {
		if p.done != nil {
			close(p.done)
		}
	})
	return nil
}

func (p *Player) Seek(position time.Duration) error {
	if p.streamer == nil {
		return fmt.Errorf("no track loaded")
	}
	posi := p.format.SampleRate.N(position)
	return p.streamer.Seek(posi)
}

func (p *Player) Duration() time.Duration {
	if p.streamer == nil {
		return 0
	}
	return p.format.SampleRate.D(p.streamer.Len())
}

func (p *Player) Position() time.Duration {
	if p.streamer == nil {
		return 0
	}
	return p.format.SampleRate.D(p.streamer.Position())
}

func (p *Player) State() PlaybackState {
	return p.state
}
