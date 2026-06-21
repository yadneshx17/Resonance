package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/yadneshx17/resonance/internal/config"
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

type model struct {
	player      *playback.Player
	library     []playback.Track
	queue       *playback.Queue
	libCursor   int
	queueCursor int
	libOffset   int
	queueOffset int
	active      string
	playingID   int
	errMsg string
	height int
}

type (
	tickMsg      time.Time
	songEndedMsg struct{ id int }
)

func Run() {
	if !config.ConfigExists() {
		config.RunSetup()
	}

	musicDir, err := config.GetMusicDir()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	q := playback.NewQueue()
	lib, err := q.ScanDir(musicDir)
	if err != nil {
		fmt.Printf("Error scanning %s: %v\n", musicDir, err)
		os.Exit(1)
	}
	if len(lib) == 0 {
		fmt.Printf("No music files found in %s\n", musicDir)
		os.Exit(1)
	}

	p := tea.NewProgram(model{
		player:  playback.NewPlayer(),
		library: lib,
		queue:   q,
		active:  "library",
	})
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v\n", err)
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
	overhead := 4 // panel borders + header + separator
	overhead += 1 // gap after panels
	overhead += 1 // controls help
	if m.player.CurrentTrack().Path != "" {
		overhead += 2 // now-playing + gap after it
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
				if m.libCursor < len(m.library)-1 {
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
			if m.active == "library" && len(m.library) > 0 {
				m.queue.Add(m.library[m.libCursor])
			}
		case "A":
			if m.active == "library" && len(m.library) > 0 {
				for _, t := range m.library {
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
		case "enter":
			if m.active == "queue" && m.queue.Len() > 0 {
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

	return fmt.Sprintf("%s %s  %s  %s  %s", icon, currtrack.Path, bar, t, volStr)
}

func (m model) buildLibBlock(slice []playback.Track, offset int) string {
	var s string
	s += headerStyle.Render("Library") + "\n"
	s += "────────\n"
	if len(slice) == 0 {
		s += "No matching tracks\n"
	} else {
		for i, t := range slice {
			idx := offset + i
			prefix := " "
			if m.active == "library" && m.libCursor == idx {
				prefix = cursorStyle.Render(">")
			}
			s += fmt.Sprintf("%s %s\n", prefix, t.Path)
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
		line := fmt.Sprintf("  %s", t.Path)
		if idx == playingIdx && m.player.State() != playback.Stopped {
			line = playingTrackStyle.Render(fmt.Sprintf("▶ %s", t.Path))
		} else if m.active == "queue" && m.queueCursor == idx {
			line = cursorStyle.Render(fmt.Sprintf("> %s", t.Path))
		}
		s += line + "\n"
	}
	return s
}

func (m model) renderColumns() string {
	vis := m.visibleRows()
	lib := m.library

	libLen := len(lib)
	if m.libOffset > libLen-vis && libLen > vis {
		m.libOffset = libLen - vis
	}
	if m.libOffset < 0 {
		m.libOffset = 0
	}

	libSlice := lib
	libOffset := 0
	if libLen > vis {
		libSlice = lib[m.libOffset : m.libOffset+vis]
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

func (m model) View() tea.View {
	var s string

	if np := m.renderNowPlaying(); np != "" {
		s += np + "\n\n"
	}

	s += m.renderColumns()
	s += "\n"

	help := "n:Next  p:Prev  Space  a:Add  A:AddAll  d:Remove  [:Vol-  ]:Vol+  m:Mute  ←→Tab:Switch  Enter:Play  q:Quit"
	if m.queue.Len() == 0 && m.player.State() == playback.Stopped {
		help = "a:Add  A:AddAll  ←→Tab:Switch  q:Quit  (queue empty)"
	}
	s += help

	if m.errMsg != "" {
		s += "\n" + m.errMsg
	}

	return tea.NewView(s)
}
