package config

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

type setupModel struct {
	state     int
	input     string
	errMsg    string
	statusMsg string
	done      bool
}

const (
	stateWelcome = iota
	stateInput
)

func RunSetup() {
	p := tea.NewProgram(setupModel{state: stateWelcome})
	if _, err := p.Run(); err != nil {
		fmt.Printf("Setup error: %v\n", err)
	}
}

func (m setupModel) Init() tea.Cmd {
	return nil
}

func (m setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch m.state {
		case stateWelcome:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "1":
				home, err := os.UserHomeDir()
				if err != nil {
					m.errMsg = "cannot determine home directory"
					return m, nil
				}
				path := home + "/Music"
				if err := ValidateMusicDir(path); err != nil {
					m.errMsg = err.Error()
					return m, nil
				}
				if err := SaveConfig(Config{MusicDir: path}); err != nil {
					m.errMsg = err.Error()
					return m, nil
				}
				m.done = true
				m.statusMsg = "Config saved! Starting Resonance..."
				return m, tea.Quit
			case "2":
				m.state = stateInput
				m.errMsg = ""
			}

		case stateInput:
			switch msg.String() {
			case "enter":
				if m.input == "" {
					m.errMsg = "Path cannot be empty"
					return m, nil
				}
				if err := ValidateMusicDir(m.input); err != nil {
					m.errMsg = err.Error()
					return m, nil
				}
				if err := SaveConfig(Config{MusicDir: m.input}); err != nil {
					m.errMsg = err.Error()
					return m, nil
				}
				m.done = true
				m.statusMsg = "Config saved! Starting Resonance..."
				return m, tea.Quit
			case "esc":
				m.state = stateWelcome
				m.input = ""
				m.errMsg = ""
			case "backspace":
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			default:
				k := msg.String()
				if len(k) == 1 {
					m.input += k
				}
			}
		}
	}
	return m, nil
}

func (m setupModel) View() tea.View {
	var s string
	s += "Welcome to Resonance 🎵\n\n"
	s += "No music library configured.\n\n"

	if m.errMsg != "" {
		s += fmt.Sprintf("Error: %s\n\n", m.errMsg)
	}

	switch m.state {
	case stateWelcome:
		s += "1. Use ~/Music\n"
		s += "2. Enter custom path\n\n"
		s += "q:Quit"

	case stateInput:
		s += "Enter music directory path:\n"
		s += "> " + m.input + "█\n\n"
		s += "Enter:Confirm  Esc:Cancel"
	}

	return tea.NewView(s)
}


