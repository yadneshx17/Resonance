package playback

import (
	"fmt"
	"time"

	"github.com/yadneshx17/resonance/internal/spotify"
	"github.com/yadneshx17/resonance/internal/types"
)

// Controller dispatches every queue track to the right backend: local tracks
// play through the beep Player, Spotify tracks are sent to an active Spotify
// Connect device via the Web API. A single queue can therefore interleave
// both sources, and the UI talks only to the controller.
type Controller struct {
	local   *Player
	spotify *spotify.Client

	deviceID string
	current  types.Track
	state    PlaybackState

	volLevel float64
	muted    bool

	spotPos       time.Duration
	spotDur       time.Duration
	spotStartedAt time.Time
}

func NewController(local *Player, sc *spotify.Client) *Controller {
	return &Controller{
		local:   local,
		spotify: sc,
		state:   Stopped,
	}
}

// PlayTrack starts playback of the given track, stopping whatever was playing
// before. If the track is a Spotify track, playback happens on the active
// Spotify Connect device.
func (c *Controller) PlayTrack(track types.Track) error {
	c.spotPos = 0
	c.spotDur = 0

	if track.Source == types.Spotify {
		c.local.Stop()
		if err := c.EnsureDevice(); err != nil {
			c.state = Stopped
			c.current = types.Track{}
			return err
		}
		if err := c.playURI(track); err != nil {
			c.state = Stopped
			c.current = types.Track{}
			return err
		}
		c.current = track
		c.state = Playing
		c.spotDur = track.Duration
		c.spotStartedAt = time.Now()
		return nil
	}

	// Hand off: stop any prior local stream (unblocks waitForSongEnd) and
	// pause remote playback so it doesn't keep blaring alongside us.
	c.local.Stop()
	c.pauseSpotify()
	if err := c.local.Load(track); err != nil {
		c.state = Stopped
		c.current = types.Track{}
		return err
	}
	if err := c.local.Play(); err != nil {
		c.state = Stopped
		c.current = types.Track{}
		return err
	}
	c.current = track
	c.state = Playing
	return nil
}

// EnsureDevice picks the device Spotify playback is routed to: the active
// device if there is one, otherwise the first available device.
func (c *Controller) EnsureDevice() error {
	if c.deviceID != "" {
		return nil
	}
	if c.spotify == nil {
		return fmt.Errorf("Spotify is not configured")
	}
	devices, err := c.spotify.GetAvailableDevices()
	if err != nil {
		return err
	}
	for _, d := range devices {
		if d.IsActive {
			c.deviceID = d.ID
			return nil
		}
	}
	for _, d := range devices {
		if d.ID != "" {
			c.deviceID = d.ID
			return nil
		}
	}
	return fmt.Errorf("no active Spotify device found - open Spotify on your phone or desktop and retry")
}

// playURI sends the track to the cached device. If that fails the cached
// device may have gone offline, so the device is re-resolved once before the
// error is surfaced.
func (c *Controller) playURI(track types.Track) error {
	err := c.spotify.PlayURI(c.deviceID, "spotify:track:"+track.ID, 0)
	if err == nil {
		return nil
	}
	c.deviceID = ""
	if err2 := c.EnsureDevice(); err2 != nil {
		return err
	}
	return c.spotify.PlayURI(c.deviceID, "spotify:track:"+track.ID, 0)
}

func (c *Controller) Pause() error {
	if c.current.Source == types.Spotify {
		if c.spotify == nil {
			return fmt.Errorf("Spotify is not configured")
		}
		if err := c.spotify.PausePlayback(c.deviceID); err != nil {
			return err
		}
		c.state = Paused
		return nil
	}
	c.state = Paused
	return c.local.Pause()
}

func (c *Controller) Resume() error {
	if c.current.Source == types.Spotify {
		if c.spotify == nil {
			return fmt.Errorf("Spotify is not configured")
		}
		if err := c.spotify.ResumePlayback(c.deviceID); err != nil {
			return err
		}
		c.state = Playing
		c.spotStartedAt = time.Now()
		return nil
	}
	c.state = Playing
	return c.local.Resume()
}

func (c *Controller) Stop() {
	c.pauseSpotify()
	c.local.Stop()
	c.state = Stopped
	c.current = types.Track{}
}

// PauseSpotify pauses any Spotify Connect playback without touching local
// state. Used on quit so music doesn't keep playing on the device.
func (c *Controller) PauseSpotify() {
	c.pauseSpotify()
}

// pauseSpotify silences whatever is playing on the active Spotify device. It
// re-resolves the device if the cached ID fails, so a handoff to local
// playback (or a quit) always stops remote music — even if it was started
// outside resonance and we never resolved a device yet.
func (c *Controller) pauseSpotify() {
	if c.spotify == nil {
		return
	}
	if c.deviceID != "" && c.spotify.PausePlayback(c.deviceID) == nil {
		return
	}
	devices, err := c.spotify.GetAvailableDevices()
	if err != nil {
		return
	}
	for _, d := range devices {
		if d.IsActive {
			c.deviceID = d.ID
			_ = c.spotify.PausePlayback(c.deviceID)
			return
		}
	}
	for _, d := range devices {
		if d.ID != "" {
			c.deviceID = d.ID
			_ = c.spotify.PausePlayback(c.deviceID)
			return
		}
	}
}

func (c *Controller) Seek(d time.Duration) error {
	if c.current.Source == types.Spotify {
		if c.spotify == nil {
			return fmt.Errorf("Spotify is not configured")
		}
		return c.spotify.SeekPlayback(c.deviceID, int(d.Milliseconds()))
	}
	return c.local.Seek(d)
}

func (c *Controller) SetVolume(delta float64) {
	c.volLevel += delta
	if c.volLevel < -3 {
		c.volLevel = -3
	}
	if c.volLevel > 3 {
		c.volLevel = 3
	}
	if c.current.Source == types.Spotify {
		if c.spotify != nil && c.deviceID != "" {
			pct := volumeToPercent(c.volLevel)
			if c.muted {
				pct = 0
			}
			_ = c.spotify.SetDeviceVolume(c.deviceID, pct)
		}
		return
	}
	c.local.SetVolume(delta)
}

func (c *Controller) Volume() float64 {
	return c.volLevel
}

func (c *Controller) Mute() {
	c.muted = true
	if c.current.Source == types.Spotify {
		if c.spotify != nil && c.deviceID != "" {
			_ = c.spotify.SetDeviceVolume(c.deviceID, 0)
		}
		return
	}
	c.local.Mute()
}

func (c *Controller) Unmute() {
	c.muted = false
	if c.current.Source == types.Spotify {
		if c.spotify != nil && c.deviceID != "" {
			_ = c.spotify.SetDeviceVolume(c.deviceID, volumeToPercent(c.volLevel))
		}
		return
	}
	c.local.Unmute()
}

func (c *Controller) IsMuted() bool {
	return c.muted
}

func (c *Controller) Position() time.Duration {
	if c.current.Source == types.Spotify {
		return c.spotPos
	}
	return c.local.Position()
}

func (c *Controller) Duration() time.Duration {
	if c.current.Source == types.Spotify {
		return c.spotDur
	}
	return c.local.Duration()
}

func (c *Controller) State() PlaybackState {
	if c.current.Source == types.Spotify {
		return c.state
	}
	return c.local.State()
}

func (c *Controller) CurrentTrack() types.Track {
	return c.current
}

// SetCurrentCover stores fetched album-art bytes on the currently playing
// track so the footer can render it.
func (c *Controller) SetCurrentCover(data []byte) {
	if c.current.Source == types.Spotify {
		c.current.CoverArt = data
	}
}

func (c *Controller) CurrentIsSpotify() bool {
	return c.current.Source == types.Spotify
}

// SpotGracePassed reports whether enough time has passed since a Spotify
// track was started for a "not playing" poll response to be trustworthy.
// It guards against auto-advancing while the device is still starting up.
func (c *Controller) SpotGracePassed(d time.Duration) bool {
	return c.current.Source == types.Spotify && time.Since(c.spotStartedAt) > d
}

// UpdateSpotifySnapshot stores the polled position/duration and reconciles
// the stored state with what the device is actually doing.
func (c *Controller) UpdateSpotifySnapshot(pos, dur time.Duration, isPlaying bool) {
	if c.current.Source != types.Spotify {
		return
	}
	c.spotPos = pos
	if dur > 0 {
		c.spotDur = dur
	}
	if isPlaying {
		c.state = Playing
		// Keep the end-detection grace fresh while real playback is
		// confirmed, so a slow device start can't be misread as "ended".
		c.spotStartedAt = time.Now()
	}
}

func volumeToPercent(vol float64) int {
	pct := int((vol + 3) / 6 * 100)
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	return pct
}
