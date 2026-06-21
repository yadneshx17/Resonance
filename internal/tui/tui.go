package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/yadneshx17/resonance/internal/config"
	"github.com/yadneshx17/resonance/internal/library"
	"github.com/yadneshx17/resonance/internal/playback"
)

var (
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(0, 1)

	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	playingIconStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00FF00"))

	pausedIconStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700"))

	playingTrackStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#20002F")).
				Background(lipgloss.Color("#F5F5DC"))

	cursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD700"))
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

const (
	setupWelcome = iota
	setupInput
)

type model struct {
	player      *playback.Player
	queue       *playback.Queue
	browser     *library.Browser
	libCursor   int
	queueCursor int
	libOffset   int
	queueOffset int
	active      string
	playingID   int
	errMsg      string
	height      int

	setup      bool
	setupState int
	setupInput string
}

type (
	tickMsg      time.Time
	songEndedMsg struct{ id int }
)

func Run() {
	m := model{
		player: playback.NewPlayer(),
		queue:  playback.NewQueue(),
		active: "library",
	}
	if !config.ConfigExists() {
		m.setup = true
		m.setupState = setupWelcome
	} else {
		musicDir, err := config.GetMusicDir()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		b, err := library.NewBrowser(musicDir)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", musicDir, err)
			os.Exit(1)
		}
		m.browser = b
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) visibleRows() int {
	if m.height == 0 {
		return 10
	}
	overhead := 4
	overhead += 1
	overhead += 1
	if m.player.CurrentTrack().Path != "" {
		overhead += 2
	}
	rows := m.height - overhead
	if rows < 1 {
		rows = 1
	}
	return rows
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height

	case tea.KeyPressMsg:
		if m.setup {
			mm, cmd := m.handleSetupKey(msg)
			return mm, cmd
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.active == "library" {
				m.active = "queue"
			} else {
				m.active = "library"
			}
		case "left":
			m.active = "library"
		case "right":
			m.active = "queue"
		case "up", "k":
			if m.active == "library" {
				if m.libCursor > 0 {
					m.libCursor--
				}
				if m.libCursor < m.libOffset {
					m.libOffset = m.libCursor
				}
			} else if m.queueCursor > 0 {
				m.queueCursor--
				if m.queueCursor < m.queueOffset {
					m.queueOffset = m.queueCursor
				}
			}
		case "down", "j":
			vis := m.visibleRows()
			if m.active == "library" {
				if m.libCursor < len(m.browser.Entries)-1 {
					m.libCursor++
				}
				if m.libCursor >= m.libOffset+vis {
					m.libOffset = m.libCursor - vis + 1
				}
			} else if m.queueCursor < m.queue.Len()-1 {
				m.queueCursor++
				if m.queueCursor >= m.queueOffset+vis {
					m.queueOffset = m.queueCursor - vis + 1
				}
			}
		case "a":
			if m.active == "library" && len(m.browser.Entries) > 0 {
				entry := m.browser.Entries[m.libCursor]
				if entry.IsDir {
					tracks, _ := m.queue.ScanDir(entry.Path)
					for _, t := range tracks {
						m.queue.Add(t)
					}
				} else {
					m.queue.Add(playback.Track{Path: entry.Path})
				}
			}
		case "A":
			if m.active == "library" {
				tracks, _ := m.queue.ScanDir(m.browser.CurrentPath)
				for _, t := range tracks {
					m.queue.Add(t)
				}
			}
		case "d":
			if m.active == "queue" && m.queue.Len() > 0 {
				m.queue.Remove(m.queueCursor)
				if m.queueCursor >= m.queue.Len() {
					m.queueCursor = max(0, m.queue.Len()-1)
				}
				if m.queueOffset >= m.queue.Len() {
					m.queueOffset = max(0, m.queue.Len()-1)
				}
			}
		case "backspace", "h":
			if m.active == "library" && m.browser.CanGoBack() {
				m.browser.GoBack()
				m.libCursor = 0
				m.libOffset = 0
			}
		case "enter":
			if m.active == "library" && len(m.browser.Entries) > 0 {
				entry := m.browser.Entries[m.libCursor]
				if entry.IsDir {
					m.browser.Open(m.libCursor)
					m.libCursor = 0
					m.libOffset = 0
				} else {
					m.playingID++
					m.errMsg = ""
					m.player.Stop()
					m.queue.Clear()
					track := playback.Track{Path: entry.Path}
					m.queue.Add(track)
					m.queue.SetCurrent(0)
					m.queueCursor = 0
					m.queueOffset = 0
					if err := m.player.Load(track); err != nil {
						m.errMsg = fmt.Sprintf("Error: %v", err)
						return m, nil
					}
					m.player.Play()
					return m, tea.Batch(waitForSongEnd(m.player, m.playingID), tick())
				}
			} else if m.active == "queue" && m.queue.Len() > 0 {
				m.playingID++
				m.errMsg = ""
				m.player.Stop()
				m.queue.SetCurrent(m.queueCursor)
				vis := m.visibleRows()
				if m.queueCursor < m.queueOffset {
					m.queueOffset = m.queueCursor
				} else if m.queueCursor >= m.queueOffset+vis {
					m.queueOffset = m.queueCursor - vis + 1
				}
				tracks := m.queue.List()
				track := tracks[m.queueCursor]
				if err := m.player.Load(track); err != nil {
					m.errMsg = fmt.Sprintf("Error: %v", err)
					return m, nil
				}
				m.player.Play()
				return m, tea.Batch(waitForSongEnd(m.player, m.playingID), tick())
			}
		case " ", "space":
			if m.player.State() == playback.Playing {
				m.player.Pause()
			} else if m.player.State() == playback.Paused {
				m.player.Resume()
			}
		case "n":
			if m.queue.Len() > 0 {
				m.playingID++
				m.errMsg = ""
				m.player.Stop()
				m.queueCursor = (m.queueCursor + 1) % m.queue.Len()
				m.queue.SetCurrent(m.queueCursor)
				vis := m.visibleRows()
				if m.queueCursor >= m.queueOffset+vis {
					m.queueOffset = m.queueCursor - vis + 1
				}
				tracks := m.queue.List()
				track := tracks[m.queueCursor]
				if err := m.player.Load(track); err != nil {
					m.errMsg = fmt.Sprintf("Error: %v", err)
					return m, nil
				}
				m.player.Play()
				return m, tea.Batch(waitForSongEnd(m.player, m.playingID), tick())
			}
		case "p":
			if m.queue.Len() > 0 {
				m.playingID++
				m.errMsg = ""
				m.player.Stop()
				m.queueCursor--
				if m.queueCursor < 0 {
					m.queueCursor = m.queue.Len() - 1
				}
				m.queue.SetCurrent(m.queueCursor)
				if m.queueCursor < m.queueOffset {
					m.queueOffset = m.queueCursor
				}
				tracks := m.queue.List()
				track := tracks[m.queueCursor]
				if err := m.player.Load(track); err != nil {
					m.errMsg = fmt.Sprintf("Error: %v", err)
					return m, nil
				}
				m.player.Play()
				return m, tea.Batch(waitForSongEnd(m.player, m.playingID), tick())
			}
		case "[":
			m.player.SetVolume(-0.1)
		case "]":
			m.player.SetVolume(0.1)
		case "m":
			if m.player.IsMuted() {
				m.player.Unmute()
			} else {
				m.player.Mute()
			}
		}

	case tickMsg:
		return m, tick()

	case songEndedMsg:
		if msg.id != m.playingID {
			return m, nil
		}
		if m.queue.Len() == 0 {
			return m, nil
		}
		m.playingID++
		m.queueCursor = (m.queueCursor + 1) % m.queue.Len()
		m.queue.SetCurrent(m.queueCursor)
		vis := m.visibleRows()
		if m.queueCursor >= m.queueOffset+vis {
			m.queueOffset = m.queueCursor - vis + 1
		}

		tracks := m.queue.List()
		track := tracks[m.queueCursor]
		m.player.Load(track)
		m.player.Play()
		return m, tea.Batch(waitForSongEnd(m.player, m.playingID), tick())
	}
	return m, nil
}

func (m model) handleSetupKey(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch m.setupState {
	case setupWelcome:
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
			m.finishSetup(path)
		case "2":
			m.setupState = setupInput
			m.errMsg = ""
		}

	case setupInput:
		switch msg.String() {
		case "enter":
			if m.setupInput == "" {
				m.errMsg = "Path cannot be empty"
				return m, nil
			}
			m.finishSetup(m.setupInput)
		case "esc":
			m.setupState = setupWelcome
			m.setupInput = ""
			m.errMsg = ""
		case "backspace":
			if len(m.setupInput) > 0 {
				m.setupInput = m.setupInput[:len(m.setupInput)-1]
			}
		default:
			k := msg.String()
			if len(k) == 1 {
				m.setupInput += k
			}
		}
	}
	return m, nil
}

func (m *model) finishSetup(path string) {
	if err := config.ValidateMusicDir(path); err != nil {
		m.errMsg = err.Error()
		return
	}
	if err := config.SaveConfig(config.Config{MusicDir: path}); err != nil {
		m.errMsg = err.Error()
		return
	}
	b, err := library.NewBrowser(path)
	if err != nil {
		m.errMsg = fmt.Sprintf("Error reading directory: %v", err)
		return
	}
	m.browser = b
	m.setup = false
	m.setupInput = ""
	m.errMsg = ""
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

func (m model) renderNowPlaying() string {
	currtrack := m.player.CurrentTrack()
	if currtrack.Path == "" {
		return ""
	}

	var icon string
	switch m.player.State() {
	case playback.Playing:
		icon = playingIconStyle.Render("▶")
	case playback.Paused:
		icon = pausedIconStyle.Render("||")
	default:
		icon = "♫"
	}

	barWidth := 20
	pos := m.player.Position()
	dur := m.player.Duration()
	bar := progressBar(pos, dur, barWidth)
	t := fmt.Sprintf("%s / %s", fmtDuration(pos), fmtDuration(dur))

	vol := m.player.Volume()
	volPct := int((vol + 3) / 6 * 100)
	if volPct < 0 {
		volPct = 0
	} else if volPct > 100 {
		volPct = 100
	}
	volStr := fmt.Sprintf("vol:%d%%", volPct)
	if m.player.IsMuted() {
		volStr = "🔇 muted"
	}

	name := currtrack.Path
	if idx := strings.LastIndexByte(currtrack.Path, '/'); idx >= 0 {
		name = currtrack.Path[idx+1:]
	}
	return fmt.Sprintf("%s %s  %s  %s  %s", icon, name, bar, t, volStr)
}

func (m model) buildLibBlock(slice []library.Entry, offset int) string {
	var s string
	s += headerStyle.Render("Library: "+m.browser.CurrentName()) + "\n"
	s += "────────────────────────\n"
	if len(slice) == 0 {
		s += "Empty\n"
	} else {
		for i, e := range slice {
			idx := offset + i
			prefix := "  "
			if m.active == "library" && m.libCursor == idx {
				prefix = cursorStyle.Render("> ")
			}
			icon := "🎵"
			if e.IsDir {
				icon = "📁"
			}
			s += fmt.Sprintf("%s%s %s\n", prefix, icon, e.Name)
		}
	}
	return s
}

func (m model) buildQueueBlock(slice []playback.Track, offset int) string {
	playingIdx := m.queue.CurrentIndex()
	var s string
	s += headerStyle.Render("Queue") + "\n"
	s += "─────\n"
	for i, t := range slice {
		idx := offset + i
		name := t.Path
		if idx := strings.LastIndexByte(t.Path, '/'); idx >= 0 {
			name = t.Path[idx+1:]
		}
		line := fmt.Sprintf("  %s", name)
		if idx == playingIdx && m.player.State() != playback.Stopped {
			line = playingTrackStyle.Render(fmt.Sprintf("▶ %s", name))
		} else if m.active == "queue" && m.queueCursor == idx {
			line = cursorStyle.Render(fmt.Sprintf("> %s", name))
		}
		s += line + "\n"
	}
	return s
}

func (m model) renderColumns() string {
	vis := m.visibleRows()
	entries := m.browser.Entries

	libLen := len(entries)
	if m.libOffset > libLen-vis && libLen > vis {
		m.libOffset = libLen - vis
	}
	if m.libOffset < 0 {
		m.libOffset = 0
	}

	libSlice := entries
	libOffset := 0
	if libLen > vis {
		libSlice = entries[m.libOffset : m.libOffset+vis]
		libOffset = m.libOffset
	}

	queueLen := m.queue.Len()
	if m.queueOffset > queueLen-vis && queueLen > vis {
		m.queueOffset = queueLen - vis
	}
	if m.queueOffset < 0 {
		m.queueOffset = 0
	}

	queueSlice := m.queue.List()
	queueOffset := 0
	if queueLen > vis {
		queueSlice = queueSlice[m.queueOffset : m.queueOffset+vis]
		queueOffset = m.queueOffset
	}

	libContent := m.buildLibBlock(libSlice, libOffset)
	queueContent := m.buildQueueBlock(queueSlice, queueOffset)

	libLines := strings.Split(libContent, "\n")
	queueLines := strings.Split(queueContent, "\n")

	maxLines := max(len(libLines), len(queueLines))
	for len(libLines) < maxLines {
		libLines = append(libLines, "")
	}
	for len(queueLines) < maxLines {
		queueLines = append(queueLines, "")
	}

	libJoined := strings.Join(libLines, "\n")
	queueJoined := strings.Join(queueLines, "\n")

	if m.active == "library" {
		libJoined = activePanelStyle.Render(libJoined)
		queueJoined = panelStyle.Render(queueJoined)
	} else {
		libJoined = panelStyle.Render(libJoined)
		queueJoined = activePanelStyle.Render(queueJoined)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, libJoined, queueJoined)
}

func (m model) setupView() tea.View {
	var s string
	s += "Welcome to Resonance 🎵\n\n"
	s += "No music library configured.\n\n"

	if m.errMsg != "" {
		s += fmt.Sprintf("Error: %s\n\n", m.errMsg)
	}

	switch m.setupState {
	case setupWelcome:
		s += "Press 1 to use ~/Music\n"
		s += "Press 2 to enter a custom path\n\n"
		s += "q:Quit"

	case setupInput:
		s += "Enter music directory path:\n"
		s += "> " + m.setupInput + "█\n\n"
		s += "Enter:Confirm  Esc:Cancel"
	}

	return tea.NewView(s)
}

func (m model) View() tea.View {
	if m.setup {
		return m.setupView()
	}
	if m.browser == nil {
		return tea.NewView("Loading...")
	}

	var s string

	if np := m.renderNowPlaying(); np != "" {
		s += np + "\n\n"
	}

	s += m.renderColumns()
	s += "\n"

	help := "n:Next  p:Prev  Space  a:Add  A:AddAll  d:Remove  [:Vol-  ]:Vol+  m:Mute  ←→Tab:Switch h/Backspace:ascend  Enter:Play  q:Quit"
	if m.queue.Len() == 0 && m.player.State() == playback.Stopped {
		help = "a:Add  A:AddAll  ←→/Tab:Switch  h/Backspace:ascend  q:Quit  (queue empty)"
	}
	s += help

	if m.errMsg != "" {
		s += "\n" + m.errMsg
	}

	return tea.NewView(s)
}
