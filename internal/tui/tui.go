package tui

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/yadneshx17/resonance/internal/playback"
)

// Holds everything visible on screen
type model struct {
	player *playback.Player
	queue  *playback.Queue
	cursor int // which queue item is selected
	// checked       int
	width, height int // terminal size
}

type (
	songEndedMsg struct{}
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
	switch msg := msg.(type) { // what is this
	case tea.KeyPressMsg:
		// fmt.Printf("KEY: %s cursor=%d\n", msg.String(), m.cursor)
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
			tracks := m.queue.List()
			if m.cursor >= len(tracks) {
				return m, nil
			}

			track := tracks[m.cursor]

			if err := m.player.Load(track); err != nil {
				return m, nil
			}

			m.player.Play()

			return m, nil
		case " ", "space":
			if m.player.State() == playback.Playing {
				m.player.Pause()
			} else {
				m.player.Resume()
			}
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	var s string
	s += "Resonance\n\nQueue: \n"
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

	s += "\n Enter:Play Space:Pause q:Quit"

	v := tea.NewView(s)
	// v.AltScreen = true // why ?

	return v
}
