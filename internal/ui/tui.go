package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/yadneshx17/resonance/internal/common"
	"github.com/yadneshx17/resonance/internal/config"
	"github.com/yadneshx17/resonance/internal/library"
	"github.com/yadneshx17/resonance/internal/playback"
	"github.com/yadneshx17/resonance/internal/types"

	"github.com/yadneshx17/resonance/internal/components"
)

var (
	headerStyle       = common.HeaderStyle
	playingIconStyle  = common.PlayingIconStyle
	pausedIconStyle   = common.PausedIconStyle
	playingTrackStyle = common.PlayingTrackStyle
	cursorStyle       = common.CursorStyle
)

const bgColor = "\x1b[48;2;26;27;56m"
const bgReset = "\x1b[49m"

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
	width       int
	tooSmall    bool

	setup      bool
	setupState int
	setupInput string

	top          *components.Top
	libraryPanel *components.Library
	visual       *components.Visual
	queuePanel   *components.Queue
	footer       *components.Footer
}

type (
	tickMsg      time.Time
	songEndedMsg struct{ id int }
)

func Run() {
	m := model{
		player:       playback.NewPlayer(),
		queue:        playback.NewQueue(),
		active:       "library",
		height:       24,
		width:        80,
		top:          &components.Top{},
		libraryPanel: &components.Library{},
		visual:       &components.Visual{},
		queuePanel:   &components.Queue{},
		footer:       &components.Footer{},
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

func (m model) fillBg(content string) string {
	w := m.width
	h := m.height
	if w < 1 {
		w = 80
	}
	if h < 1 {
		h = 24
	}
	lines := strings.Split(content, "\n")
	out := make([]string, 0, h)
	for _, line := range lines {
		lw := lipgloss.Width(line)
		if lw < w {
			line += strings.Repeat(" ", w-lw)
		}
		out = append(out, bgColor+line+bgReset)
	}
	for len(out) < h {
		out = append(out, bgColor+strings.Repeat(" ", w)+bgReset)
	}
	return strings.Join(out, "\n")
}

func (m model) visibleRows() int {
	if m.height == 0 {
		return 20
	}
	// mainHeight = m.height - top(1) - playback(5)
	// inside border = mainHeight - 2
	// subtract 2 header lines (title + separator)
	// = m.height - 6 - 2 - 2 = m.height - 10
	n := m.height - 10
	if n < 0 {
		return 0
	}
	return n
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.tooSmall = m.width < 60 || m.height < 24

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
					m.queue.Add(types.Track{Path: entry.Path})
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
		case "esc", "h":
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
					track := types.Track{Path: entry.Path}
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

	return tea.View{
		AltScreen: true,
		Content:   m.fillBg(s),
	}
}

func (m model) View() tea.View {
	if m.setup {
		return m.setupView()
	}

	if m.tooSmall {
		badStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#d46159")).Bold(true)

		wStr := fmt.Sprintf("%d", m.width)
		if m.width < 60 {
			wStr = badStyle.Render(fmt.Sprintf("%d", m.width))
		}
		hStr := fmt.Sprintf("%d", m.height)
		if m.height < 24 {
			hStr = badStyle.Render(fmt.Sprintf("%d", m.height))
		}

		s := fmt.Sprintf("Terminal size too small\nWidth = %s  Height = %s\n\nNeeded for current config\nWidth >= 60, Height >= 24", wStr, hStr)
		placed := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s)
		return tea.View{
			AltScreen: true,
			Content:   m.fillBg(placed),
		}
	}

	topHeight := 1
	playbackHeight := 5
	mainHeight := m.height - topHeight - playbackHeight

	// gap := 1
	sideWidth := m.width / 4
	visWidth := m.width / 2
	queueWidth := m.width - sideWidth - visWidth

	vis := m.visibleRows()
	entries := m.browser.Entries
	libSlice := entries
	libOffset := 0
	if len(entries) > vis {
		off := m.libOffset
		if off > len(entries)-vis {
			off = len(entries) - vis
		}
		if off < 0 {
			off = 0
		}
		libSlice = entries[off : off+vis]
		libOffset = off
	}
	m.libraryPanel.SetData(components.LibData{
		Entries: libSlice,
		Title:   m.browser.CurrentName(),
		Cursor:  m.libCursor,
		Offset:  libOffset,
		Total:   len(m.browser.Entries),
		Active:  m.active == "library",
		Width:   sideWidth,
		Height:  mainHeight,
	})

	tracks := m.queue.List()
	queueSlice := tracks
	queueOffset := 0
	if len(tracks) > vis {
		off := m.queueOffset
		if off > len(tracks)-vis {
			off = len(tracks) - vis
		}
		if off < 0 {
			off = 0
		}
		queueSlice = tracks[off : off+vis]
		queueOffset = off
	}
	m.queuePanel.SetData(components.QueueData{
		Tracks:     queueSlice,
		PlayingIdx: m.queue.CurrentIndex(),
		Playing:    m.player.State() != playback.Stopped,
		Cursor:     m.queueCursor,
		Offset:     queueOffset,
		Total:      m.queue.Len(),
		Active:     m.active == "queue",
		Width:      queueWidth,
		Height:     mainHeight,
	})

	track := m.player.CurrentTrack()
	trackName := ""
	if track.Path != "" {
		if idx := strings.LastIndexByte(track.Path, '/'); idx >= 0 {
			trackName = track.Path[idx+1:]
		}
	}
	var stateIcon string
	switch m.player.State() {
	case playback.Playing:
		stateIcon = 	playingIconStyle.Render(common.Play)
	case playback.Paused:
		stateIcon = 		pausedIconStyle.Render(common.Pause)
	default:
		stateIcon = common.MusicNote
	}

	m.visual.SetData(components.VisualData{
		TrackName: trackName,
		StateIcon: stateIcon,
		Position:  m.player.Position(),
		Duration:  m.player.Duration(),
		VolLevel:  m.player.Volume(),
		Muted:     m.player.IsMuted(),
		Height:    mainHeight,
		Width:     visWidth,
	})

	m.footer.SetData(components.FooterData{
		TrackName: trackName,
		StateIcon: stateIcon,
		Position:  m.player.Position(),
		Duration:  m.player.Duration(),
		VolLevel:  m.player.Volume(),
		Muted:     m.player.IsMuted(),
		Height:    playbackHeight,
		Width:     m.width,
	})

	m.top.SetTrackCount(m.browser.TrackCount())

	s := lipgloss.JoinVertical(
		lipgloss.Left,
		m.top.View(topHeight, m.width),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.libraryPanel.View(),
			m.visual.View(),
			m.queuePanel.View(),
		),
		m.footer.View(),
	)

	return tea.View{
		AltScreen: true,
		Content:   m.fillBg(s),
	}
}
