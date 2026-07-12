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
	playingIconStyle  = common.PlayingIconStyle
	pausedIconStyle   = common.PausedIconStyle
	playingTrackStyle = common.PlayingTrackStyle
	cursorStyle       = common.CursorStyle
)

// const bgColor = "\x1b[48;2;26;27;56m" // background sequence
const bgColor = "\x1b[48;2;30;30;46m"

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

	// Setup
	setup      bool
	setupState int
	setupInput string

	// Componnets
	top          *components.Top
	libraryPanel *components.Library
	visual       *components.Visual
	queuePanel   *components.Queue
	footer       *components.Footer

	// Search
	searchMode  bool
	searchQuery string
	searchIdx   []int

	// Help
	showHelp bool
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
		line = strings.ReplaceAll(line, "\x1b[m", "\x1b[m"+bgColor)
		line = strings.ReplaceAll(line, "\x1b[0m", "\x1b[0m"+bgColor)
		lw := lipgloss.Width(line)
		if lw < w {
			line += strings.Repeat(" ", w-lw)
		}
		out = append(out, bgColor+line)
	}
	for len(out) < h {
		out = append(out, bgColor+strings.Repeat(" ", w))
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
		m.tooSmall = m.width < 84 || m.height < 24

	case tea.KeyPressMsg:
		if m.setup {
			mm, cmd := m.handleSetupKey(msg)
			return mm, cmd
		}

		if m.showHelp {
			switch msg.String() {
			case "?", "esc", "q":
				m.showHelp = false
			}
			return m, nil
		}

		// Search mode intercepts printable keys first
		if m.searchMode {
			switch msg.String() {
			case "esc":
				m.searchMode = false
				m.searchQuery = ""
				return m, nil
			case "backspace":
				if len(m.searchQuery) > 0 {
					runes := []rune(m.searchQuery)
					m.searchQuery = string(runes[:len(runes)-1])
					m.rebuildSearch()
				}
				return m, nil
			default:
				if len(msg.String()) == 1 {
					m.searchQuery += msg.String()
					m.rebuildSearch()
					return m, nil
				}
			}
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
			m.searchQuery = ""
		case "right":
			m.active = "queue"
			m.searchQuery = ""
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
				maxLen := len(m.browser.Entries)
				if m.searchMode {
					maxLen = len(m.searchIdx)
				}
				if m.libCursor < maxLen-1 {
					m.libCursor++
				}
				if m.libCursor >= m.libOffset+vis {
					m.libOffset = m.libCursor - vis + 1
				}
			} else if m.active == "queue" {
				maxLen := m.queue.Len()
				if m.searchMode {
					maxLen = len(m.searchIdx)
				}
				if m.queueCursor < maxLen-1 {
					m.queueCursor++
				}
				if m.queueCursor >= m.queueOffset+vis {
					m.queueOffset = m.queueCursor - vis + 1
				}
			}
		case "a":
			if m.active == "library" && len(m.browser.Entries) > 0 {
				idx := m.libCursor
				if m.searchMode && idx < len(m.searchIdx) {
					idx = m.searchIdx[idx]
				}
				entry := m.browser.Entries[idx]
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
				idx := m.queueCursor
				if m.searchMode && idx < len(m.searchIdx) {
					idx = m.searchIdx[idx]
				}
				m.queue.Remove(idx)
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
				idx := m.libCursor
				if m.searchMode && idx < len(m.searchIdx) {
					idx = m.searchIdx[idx]
				}
				entry := m.browser.Entries[idx]
				if entry.IsDir {
					m.browser.Open(idx)
					m.libCursor = 0
					m.libOffset = 0
					m.searchMode = false
					m.searchQuery = ""
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
				idx := m.queueCursor
				if m.searchMode && idx < len(m.searchIdx) {
					idx = m.searchIdx[idx]
				}
				m.playingID++
				m.errMsg = ""
				m.player.Stop()
				m.queue.SetCurrent(idx)
				vis := m.visibleRows()
				if m.queueCursor < m.queueOffset {
					m.queueOffset = m.queueCursor
				} else if m.queueCursor >= m.queueOffset+vis {
					m.queueOffset = m.queueCursor - vis + 1
				}
				tracks := m.queue.List()
				track := tracks[idx]
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
		case "/":
			if !m.searchMode {
				m.searchMode = true
				m.searchQuery = ""
				return m, nil
			}
		case "<":
			if m.player.State() != playback.Stopped {
				pos := m.player.Position()
				dur := m.player.Duration()
				newPos := pos - 5*time.Second
				if newPos < 0 {
					newPos = dur
				}
				m.player.Seek(newPos)
			}
		case ">":
			if m.player.State() != playback.Stopped {
				pos := m.player.Position()
				dur := m.player.Duration()
				newPos := pos + 5*time.Second
				if newPos > dur {
					newPos = dur
				}
				m.player.Seek(newPos)
			}
		case "?":
			m.showHelp = !m.showHelp
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

func (m *model) rebuildSearch() {
	if m.searchQuery == "" {
		m.searchIdx = nil
		return
	}
	var src []string // names from either browser enteries or queue strings
	if m.active == "library" {
		for _, e := range m.browser.Entries {
			src = append(src, e.Name)
		}
	} else {
		for _, t := range m.queue.List() {
			name := t.Path
			if slash := strings.LastIndexByte(t.Path, '/'); slash >= 0 {
				name = t.Path[slash+1:]
			}
			src = append(src, name)
		}
	}

	m.searchIdx = nil
	for i, name := range src {
		if common.FuzzyMatch(m.searchQuery, name) {
			m.searchIdx = append(m.searchIdx, i)
		}
	}

	if len(m.searchIdx) == 0 {
		return
	}

	// clamping of cursor
	if m.active == "library" && m.libCursor >= len(m.searchIdx) {
		m.libCursor = len(m.searchIdx) - 1
	}
	if m.active == "queue" && m.queueCursor >= len(m.searchIdx) {
		m.queueCursor = len(m.searchIdx) - 1
	}
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
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#B4BEFE"))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CDD6F4"))
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F38BA8"))
	optionStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#45475A")).
		Padding(0, 2).
		Foreground(lipgloss.Color("#CDD6F4"))
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#89B4FA")).
		Padding(0, 1).
		Bold(true).
		Foreground(lipgloss.Color("#CDD6F4")).
		Width(50)

	var s string
	s += titleStyle.Render("Welcome to Resonance") + "\n\n"
	s += normalStyle.Render("No music library configured.") + "\n\n"

	if m.errMsg != "" {
		s += errorStyle.Render(common.Error+" "+m.errMsg) + "\n\n"
	}

	switch m.setupState {
	case setupWelcome:
		opt1 := optionStyle.Render("[1]  " + common.Directory + "  ~/Music")
		opt2 := optionStyle.Render("[2]  " + common.Search + "  Custom path")
		s += opt1 + "\n" + opt2 + "\n\n"
		s += keyStyle.Render("  q  quit")

	case setupInput:
		s += normalStyle.Render("Enter music directory path:") + "\n\n"
		inputLine := "> " + m.setupInput + "█"
		s += inputStyle.Render(inputLine) + "\n\n"
		s += keyStyle.Render("  " + common.Confirm + " confirm    " + common.Cancel + " cancel")
	}

	box := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s)
	return tea.View{
		AltScreen: true,
		Content:   m.fillBg(box),
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

		s := fmt.Sprintf("Terminal size too small\nWidth = %s  Height = %s\n\nNeeded for current config\nWidth >= 84, Height >= 24", wStr, hStr)
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
	if m.searchMode && m.active == "library" && len(m.searchIdx) < len(entries) {
		filtered := make([]types.Entry, len(m.searchIdx))
		for i, idx := range m.searchIdx {
			filtered[i] = entries[idx]
		}
		entries = filtered
	}
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
		Entries:     libSlice,
		Title:       m.browser.CurrentName(),
		Cursor:      m.libCursor,
		Offset:      libOffset,
		Total:       len(m.browser.Entries),
		Active:      m.active == "library",
		Width:       sideWidth,
		Height:      mainHeight,
		SearchMode:  m.searchMode && m.active == "library",
		SearchQuery: m.searchQuery,
	})

	// queue
	tracks := m.queue.List()
	if m.searchMode && m.active == "queue" && len(m.searchIdx) < len(tracks) {
		filtered := make([]types.Track, len(m.searchIdx))
		for i, idx := range m.searchIdx {
			filtered[i] = tracks[idx]
		}
		tracks = filtered
	}
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
		Tracks:      queueSlice,
		PlayingIdx:  m.queue.CurrentIndex(),
		Playing:     m.player.State() != playback.Stopped,
		Cursor:      m.queueCursor,
		Offset:      queueOffset,
		Total:       m.queue.Len(),
		Active:      m.active == "queue",
		Width:       queueWidth,
		Height:      mainHeight,
		SearchMode:  m.searchMode && m.active == "queue",
		SearchQuery: m.searchQuery,
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
		stateIcon = playingIconStyle.Render(common.Play)
	case playback.Paused:
		stateIcon = pausedIconStyle.Render(common.Pause)
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

	if m.showHelp {
		s = m.helpView()
	}

	return tea.View{
		AltScreen: true,
		Content:   m.fillBg(s),
	}
}

func (m model) helpView() string {
	helpW := 54
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#B4BEFE"))
	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render("│")
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#89B4FA"))

	groups := []struct {
		section string
		items   [][2]string
	}{
		{"Navigation", [][2]string{
			{"↑ / k , ↓ / j", "Move cursor"},
			{"Tab  ←  →", "Switch panel"},
			{"Enter (dir)", "Open directory"},
			{"Enter (track)", "Play track"},
			{"Backspace / h", "Parent directory"},
		}},
		{"Search", [][2]string{
			{"/", "Enter search mode"},
			{"Esc", "Exit search"},
			{"Backspace", "Delete character"},
		}},
		{"Queue", [][2]string{
			{"a (track)", "Add to queue"},
			{"a (dir)", "Add all from dir"},
			{"A", "Add all current dir"},
			{"d", "Remove from queue"},
		}},
		{"Playback", [][2]string{
			{"Space", "Pause / Resume"},
			{"n , p", "Next / Previous track"},
			{"[ , ]", "Volume down / up"},
			{"m", "Mute / Unmute"},
			{"< , >", "Seek -5s / +5s"},
		}},
		{"General", [][2]string{
			{"q / Ctrl+C", "Quit"},
			{"?", "Close this help"},
		}},
	}

	// find the longest key across all groups
	maxKeyW := 0
	for _, g := range groups {
		for _, item := range g.items {
			if w := lipgloss.Width(item[0]); w > maxKeyW {
				maxKeyW = w
			}
		}
	}

	var b strings.Builder
	for _, g := range groups {
		b.WriteString(header.Render(g.section) + "\n")
		for _, item := range g.items {
			key := item[0]
			padded := key + strings.Repeat(" ", maxKeyW-lipgloss.Width(key))
			line := fmt.Sprintf("  %s  %s  %s", keyStyle.Render(padded), sep, item[1])
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(dim.Render("  ? or Esc to close"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#45475A")).
		Padding(1, 3).
		Width(helpW).
		Render(b.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
