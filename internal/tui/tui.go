package tui

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/yadneshx17/resonance/internal/playback"
)

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", m, s)
}

func progressBar(pos, dur time.Duration, width int) string {
	if dur == 0 {
		return ""
	}
	filled := int(float64(pos) / float64(dur) * float64(width))
	if filled > width {
		filled = width
	}
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}

// Holds everything visible on screen
type model struct {
	player        *playback.Player
	queue         *playback.Queue
	cursor        int // which queue item is selected
	playingID     int // increments on each play, filters stale songEndedMsg
	width, height int
}

type (
	tickMsg      time.Time
	songEndedMsg struct{ id int }
)

func Run() {
	q := playback.NewQueue()
	q.PopulateQueue("Music")

	p := tea.NewProgram(model{
		player: playback.NewPlayer(),
		queue:  q,
	})
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

// runs after model is created.
func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < m.queue.Len()-1 {
				m.cursor++
			}
		case "enter":
			m.playingID++
			tracks := m.queue.List()
			if m.cursor >= len(tracks) {
				return m, nil
			}
			track := tracks[m.cursor]
			if err := m.player.Load(track); err != nil {
				return m, nil
			}
			m.player.Play()
			return m, tea.Batch(waitForSongEnd(m.player, m.playingID), tick())
		case " ", "space":
			if m.player.State() == playback.Playing {
				m.player.Pause()
			} else {
				m.player.Resume()
			}
		case "n":
			m.playingID++
			m.player.Stop()
			m.cursor = (m.cursor + 1) % m.queue.Len()
			tracks := m.queue.List()
			track := tracks[m.cursor]
			m.player.Load(track)
			m.player.Play()
			return m, tea.Batch(waitForSongEnd(m.player, m.playingID), tick())
		case "p":
			m.playingID++
			m.player.Stop()
			m.cursor--
			if m.cursor < 0 {
				m.cursor = m.queue.Len() - 1
			}
			tracks := m.queue.List()
			track := tracks[m.cursor]
			m.player.Load(track)
			m.player.Play()
			return m, tea.Batch(waitForSongEnd(m.player, m.playingID), tick())
		}

	case tickMsg:
		return m, tick()

	case songEndedMsg:
		if msg.id != m.playingID {
			return m, nil
		}
		m.playingID++
		m.cursor = (m.cursor + 1) % m.queue.Len()

		// plays next song
		tracks := m.queue.List()
		track := tracks[m.cursor]
		m.player.Load(track)
		m.player.Play()
		return m, tea.Batch(waitForSongEnd(m.player, m.playingID), tick())
	}
	return m, nil
}

func waitForSongEnd(player *playback.Player, id int) tea.Cmd {
	return func() tea.Msg {
		player.Wait()
		return songEndedMsg{id: id}
	}
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) View() tea.View {
	var s string
	s += "\nResonance\n\n"

	// Now Playing section
	track := m.player.CurrentTrack()
	if track.Path != "" {
		switch m.player.State() {
		case playback.Paused:
			s += "|| "
		default:
			s += "♫ "

		}
		barWidth := 20
		pos := m.player.Position()
		dur := m.player.Duration()
		s += fmt.Sprintf("%s %s / %s\n", progressBar(pos, dur, barWidth), fmtDuration(pos), fmtDuration(dur))
	}

	s += "Queue:\n"
	if m.queue.Len() > 0 {
		if m.cursor >= m.queue.Len() {
			m.cursor = m.queue.Len() - 1
		}

		if m.cursor < 0 {
			m.cursor = 0
		}
	}
	for i, t := range m.queue.List() {
		c := " "
		if m.cursor == i {
			c = ">"
		}

		s += fmt.Sprintf(" %s %s\n", c, t.Path)
	}

	s += "\n\n Enter:Play Space:Pause q:Quit n:Next p:Previous"

	v := tea.NewView(s)
	// v.AltScreen = true // why ?

	return v
}
