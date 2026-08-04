package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/yadneshx17/resonance/internal/common"
	"github.com/yadneshx17/resonance/internal/config"
	"github.com/yadneshx17/resonance/internal/library"
	"github.com/yadneshx17/resonance/internal/playback"
	"github.com/yadneshx17/resonance/internal/spotify"
	"github.com/yadneshx17/resonance/internal/types"

	"github.com/yadneshx17/resonance/internal/components"
)

var (
	playingIconStyle  = common.PlayingIconStyle
	pausedIconStyle   = common.PausedIconStyle
	playingTrackStyle = common.PlayingTrackStyle
	cursorStyle       = common.CursorStyle
)

const bgColor = "\x1b[48;2;30;30;46m"

// const bgColor = "\x1b[0;49m" // transparent

const (
	setupWelcome = iota
	setupInput
)

type model struct {
	player        *playback.Player
	controller    *playback.Controller
	queue         *playback.Queue
	browser       *library.Browser
	libCursor     int
	tracksCursor  int
	queueCursor   int
	libOffset     int
	tracksOffset  int
	queueOffset   int
	active        string
	playingID     int
	errMsg        string
	spotPollCount int // throttle /v1/me/player polls to every 5th tick
	height        int
	width         int
	tooSmall      bool
	selectedDir   string
	fileEntries   []types.Entry
	subdirCount   int

	// Setup
	setup      bool
	setupState int
	setupInput string

	// Componnets
	top               *components.Top
	libraryPanel      *components.Library
	tracksPanel       *components.Tracks
	queuePanel        *components.Queue
	footer            *components.Footer
	sourceSwitchPanel *components.SourceSwitch

	// Search
	searchMode  bool
	searchQuery string
	searchIdx   []int

	// Help
	showHelp bool

	// Source Switching
	source       string // "local" | "spotify"
	sourceSwitch bool
	sourceCursor int // overlay selection (0=local, 1=spotify)

	// Spotify browsing
	spotifyClient   *spotify.Client
	spotifyLoggedIn bool

	spotifyCategory  int // 0=Liked Songs, 1=Playlists, 2=Albums
	spotifyItems     []types.Track
	spotifyPlaylists []spotify.SpotifyPlaylist
	spotifyAlbums    []spotify.SavedAlbum
	spotifyTotal     int
	spotifyOffset    int
	spotifyHasMore   bool
	spotifyLoading   bool
	spotifyErr       string
	spotifyCursor    int // library category cursor (0-2)
	spotifyScroll    int // content panel scroll

	// API total of the currently loaded playlists/albums list. Kept separate
	// from spotifyTotal because a drill overwrites spotifyTotal with the
	// drill's track total; this lets us restore the list state on exit.
	spotifyListTotal int

	// Spotify drill-down (playlists / albums)
	spotifyDrillType string // "" | "playlist" | "album"
	spotifyDrillID   string // playlist or album ID for refetches
	spotifyDrillName string // content title when drilled

	// Spotify content panel cursor (independent of library cursor)
	spotifyContentCursor int

	// Cursor/scroll position saved when drilling into a playlist/album and
	// restored on backspace, so the list returns to where the user left.
	spotifyContentCursorBackup int
	spotifyScrollBackup        int
}

// Messages
type (
	tickMsg      time.Time
	songEndedMsg struct{ id int }
)

type spotifyTracksMsg struct {
	items    []types.Track
	total    int
	offset   int
	category int // stale-response guard (M5)
}

type spotifyPlaylistsMsg struct {
	playlists []spotify.SpotifyPlaylist
	total     int
	offset    int
	category  int // stale-response guard (M5)
}

type spotifyPlaylistTracksMsg struct {
	items  []types.Track
	total  int
	offset int
	name   string // name of the playlist
}

type spotifyAlbumsMsg struct {
	albums   []spotify.SavedAlbum
	total    int
	offset   int
	category int // stale-response guard (M5)
}

type spotifyAlbumTracksMsg struct {
	items  []types.Track
	total  int
	offset int
	name   string // name of the album
}

type spotifyErrorMsg struct {
	err string
}

type spotifyPollMsg struct {
	state *spotify.PlayerState
}

func Run() {
	m := model{
		player:            playback.NewPlayer(),
		queue:             playback.NewQueue(),
		active:            "library",
		source:            "local",
		height:            24,
		width:             80,
		top:               &components.Top{},
		libraryPanel:      &components.Library{},
		tracksPanel:       &components.Tracks{},
		queuePanel:        &components.Queue{},
		footer:            &components.Footer{},
		sourceSwitchPanel: &components.SourceSwitch{},
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
		m.selectDirAtCursor()
	}

	if spotify.CredentialsExist() {
		cred, err := spotify.LoadCredentials()
		if err != nil {
			fmt.Printf("Error Loading Cred: %v\n", err)
			os.Exit(1)
		}

		m.spotifyClient = spotify.NewClient(cred)
		m.spotifyLoggedIn = true
	}

	m.controller = playback.NewController(m.player, m.spotifyClient)

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

func (m model) browserDirs() []types.Entry {
	var dirs []types.Entry

	// Add dot entry if current folder has songs
	if m.browser.HasDirectSongs() {
		dirs = append(dirs, types.Entry{
			Name:  "•",
			Path:  m.browser.CurrentPath,
			IsDir: false,
		})
	}

	for _, e := range m.browser.Entries {
		if e.IsDir {
			dirs = append(dirs, e)
		}
	}
	return dirs
}

func (m model) trackFileEntries() []types.Entry {
	if m.fileEntries != nil {
		return m.fileEntries
	}
	return m.browserFileEntries()
}

func (m model) browserFileEntries() []types.Entry {
	var files []types.Entry
	for _, e := range m.browser.Entries {
		if !e.IsDir {
			files = append(files, e)
		}
	}
	return files
}

func loadDirFiles(dir string) []types.Entry {
	items, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var entries []types.Entry
	for _, item := range items {
		if item.IsDir() || !strings.HasSuffix(strings.ToLower(item.Name()), ".mp3") {
			continue
		}
		fullPath := filepath.Join(dir, item.Name())
		meta, _ := common.ReadMetadata(fullPath)
		entries = append(entries, types.Entry{
			Name:     item.Name(),
			Path:     fullPath,
			Title:    meta.Title,
			Artist:   meta.Artist,
			Album:    meta.Album,
			Duration: meta.Duration,
			CoverArt: meta.CoverArt,
		})
	}
	return entries
}

func (m model) libIdxFromCursor(dirsIdx int) int {
	dirs := m.browserDirs()
	if dirsIdx >= 0 && dirsIdx < len(dirs) {
		entry := dirs[dirsIdx]
		if entry.Name == "•" {
			return -1 // sentinel for dot entry
		}
		for i, e := range m.browser.Entries {
			if e.Path == entry.Path {
				return i
			}
		}
	}
	return 0
}

func (m *model) selectDirAtCursor() {
	dirs := m.browserDirs()
	idx := m.libCursor
	if m.searchMode && m.active == "library" && m.libCursor < len(m.searchIdx) {
		idx = m.searchIdx[m.libCursor]
	}
	m.subdirCount = 0
	if idx < len(dirs) {
		sel := dirs[idx]
		if sel.Name == "•" {
			m.selectedDir = m.browser.CurrentPath
			m.fileEntries = m.browser.DirectSongs()
		} else {
			m.selectedDir = sel.Path
			m.fileEntries = loadDirFiles(sel.Path)
			if len(m.fileEntries) == 0 {
				items, err := os.ReadDir(sel.Path)
				if err == nil {
					for _, item := range items {
						if item.IsDir() && !strings.HasPrefix(item.Name(), ".") {
							m.subdirCount++
						}
					}
				}
			}
		}
		m.tracksCursor = 0
		m.tracksOffset = 0
	}
}

func (m model) dirTrackCount() int {
	return len(m.trackFileEntries())
}

// contentHeight returns the height of the main layout area: the full window
// minus the top bar (1 line) and footer (11 lines).
func (m model) contentHeight() int {
	n := m.height - 12
	if n < 0 {
		return 0
	}
	return n
}

// libHeight / libVis mirror the library panel sizing in buildMainView.
func (m model) libHeight() int {
	n := m.contentHeight() * 30 / 100
	if n < 3 {
		return 3
	}
	return n
}

func (m model) libVis() int {
	return m.libHeight() - 2
}

func (m model) queueVis() int {
	n := m.contentHeight() - m.libHeight() - 2
	if n < 0 {
		return 0
	}
	return n
}

// trackVis returns the number of data rows that fit in the tracks/content
// panel. The panel is contentHeight tall; two rows are its borders and two
// more are the header + separator line, so the data area is contentHeight-4.
// Keeping this identical to buildMainView's slicing is what guarantees the
// cursor can never scroll below the visible area.
func (m model) trackVis() int {
	n := m.contentHeight() - 4
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
		m.tooSmall = m.width < 84 || m.height < 28

	case tea.KeyPressMsg:
		if m.setup {
			mm, cmd := m.handleSetupKey(msg)
			return mm, cmd
		}

		if m.showHelp {
			switch msg.String() {
			case "?", "esc", "q":
				m.showHelp = false
			case "ctrl+c":
				return m, m.quit()
			}
			return m, nil
		}

		if m.sourceSwitch {
			switch msg.String() {
			case "up", "k":
				if m.sourceCursor > 0 {
					m.sourceCursor--
				}
			case "down", "j":
				if m.sourceCursor < 1 {
					m.sourceCursor++
				}
			case "enter":
				if m.sourceCursor == 1 && !m.spotifyLoggedIn {
					// can't switch to spotify if not logged in
					// some kind of actionable indicator for USER
				} else {
					m.source = []string{"local", "spotify"}[m.sourceCursor]
					m.sourceSwitch = false
					if m.source == "spotify" {
						m.spotifyCursor = 0
						m.spotifyItems = nil
						m.spotifyPlaylists = nil
						m.spotifyAlbums = nil
						m.spotifyDrillType = ""
						m.spotifyDrillID = ""
						m.spotifyDrillName = ""
						m.spotifyContentCursor = 0
						m.spotifyScroll = 0
						return m.loadCategory(0)
					}
				}
			case "esc", "ctrl+s":
				m.sourceSwitch = false

			case "ctrl+c":
				return m, m.quit()
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
			return m, m.quit()
		case "tab":
			switch m.active {
			case "library":
				m.active = "tracks"
				m.searchMode = false
				m.searchQuery = ""
			case "tracks":
				m.active = "queue"
				m.searchMode = false
				m.searchQuery = ""
			default:
				m.active = "library"
				m.searchMode = false
				m.searchQuery = ""
			}
		case "left":
			// used by l key for directory navigation
		case "right":
			// reserved
		case "up", "k":
			if m.source == "spotify" && (m.active == "library" || m.active == "tracks") {
				return m.handleSpotifyUp()
			}

			if m.active == "library" {
				if m.libCursor > 0 {
					m.libCursor--
				}
				if m.libCursor < m.libOffset {
					m.libOffset = m.libCursor
				}
				m.selectDirAtCursor()
			} else if m.active == "tracks" {
				if m.tracksCursor > 0 {
					m.tracksCursor--
				}
				if m.tracksCursor < m.tracksOffset {
					m.tracksOffset = m.tracksCursor
				}
			} else if m.active == "queue" && m.queueCursor > 0 {
				m.queueCursor--
				if m.queueCursor < m.queueOffset {
					m.queueOffset = m.queueCursor
				}
			}
		case "down", "j":
			if m.source == "spotify" && (m.active == "library" || m.active == "tracks") {
				return m.handleSpotifyDown()
			}

			libVis := m.libVis()
			queueVis := m.queueVis()
			trackVis := m.trackVis()
			if m.active == "library" {
				maxLen := len(m.browserDirs())
				if m.searchMode {
					maxLen = len(m.searchIdx)
				}
				if m.libCursor < maxLen-1 {
					m.libCursor++
				}
				if m.libCursor >= m.libOffset+libVis {
					m.libOffset = m.libCursor - libVis + 1
				}
				m.selectDirAtCursor()
			} else if m.active == "tracks" {
				maxLen := m.dirTrackCount()
				if m.searchMode {
					maxLen = len(m.searchIdx)
				}
				if m.tracksCursor < maxLen-1 {
					m.tracksCursor++
				}
				if m.tracksCursor >= m.tracksOffset+trackVis {
					m.tracksOffset = m.tracksCursor - trackVis + 1
				}
			} else if m.active == "queue" {
				maxLen := m.queue.Len()
				if m.searchMode {
					maxLen = len(m.searchIdx)
				}
				if m.queueCursor < maxLen-1 {
					m.queueCursor++
				}
				if m.queueCursor >= m.queueOffset+queueVis {
					m.queueOffset = m.queueCursor - queueVis + 1
				}
			}
		case "a":
			if m.source == "spotify" {
				return m.handleSpotifySmallA()
			}

			if m.active == "library" {
				dirs := m.browserDirs()
				idx := m.libCursor
				if m.searchMode && idx < len(m.searchIdx) {
					idx = m.searchIdx[idx]
				}
				if idx < len(dirs) {
					sel := dirs[idx]
					if sel.Name == "•" {
						for _, e := range m.browser.DirectSongs() {
							m.queue.Add(types.Track{
								Path:     e.Path,
								Title:    e.Title,
								Artist:   e.Artist,
								Album:    e.Album,
								CoverArt: e.CoverArt,
								Duration: e.Duration,
								Source:   types.Local,
							})
						}
					} else {
						tracks, _ := m.queue.ScanDir(sel.Path)
						for _, t := range tracks {
							m.queue.Add(t)
						}
					}
				}
			} else if m.active == "tracks" {
				entries := m.trackFileEntries()
				idx := m.tracksCursor
				if m.searchMode && idx < len(m.searchIdx) {
					idx = m.searchIdx[idx]
				}
				if idx < len(entries) {
					e := entries[idx]
					track := types.Track{
						Path:     e.Path,
						Title:    e.Title,
						Artist:   e.Artist,
						Album:    e.Album,
						CoverArt: e.CoverArt,
						Duration: e.Duration,
						Source:   types.Local,
					}
					m.queue.Add(track)
				}
			}
		case "A":
			if m.active == "library" || m.active == "tracks" {
				dir := m.selectedDir
				if dir == "" {
					dir = m.browser.CurrentPath
				}
				tracks, _ := m.queue.ScanDir(dir)
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
		case "backspace", "h":
			if m.source == "spotify" {
				return m.handleSpotifyBackspace()
			}

			if m.active == "library" {
				m.selectedDir = ""
				m.fileEntries = nil
				m.tracksCursor = 0
				m.tracksOffset = 0
				if m.browser.CanGoBack() {
					m.browser.GoBack()
					m.libCursor = 0
					m.libOffset = 0
				}
				if len(m.browserDirs()) > 0 {
					m.selectDirAtCursor()
				}
			}
		case "enter":
			if m.source == "spotify" && m.active == "tracks" {
				return m.handleSpotifyContentEnter()
			}
			if m.source == "spotify" && m.active == "library" {
				// library panel: content already hover-loaded, nothing to do
				return m, nil
			}

			if m.active == "library" {
				dirs := m.browserDirs()
				idx := m.libCursor
				if m.searchMode && idx < len(m.searchIdx) {
					idx = m.searchIdx[idx]
				}
				if idx < len(dirs) {
					sel := dirs[idx]
					if sel.Name != "•" && sel.IsDir {
						// Enter directory (drill-down)
						m.browser.Open(m.libIdxFromCursor(idx))
						m.libCursor = 0
						m.libOffset = 0
						m.selectedDir = ""
						m.fileEntries = nil
						m.tracksCursor = 0
						m.tracksOffset = 0
						m.searchMode = false
						m.searchQuery = ""
						if len(m.browserDirs()) > 0 {
							m.selectDirAtCursor()
						}
					}
				}
			} else if m.active == "tracks" {
				entries := m.trackFileEntries()
				idx := m.tracksCursor
				if m.searchMode && idx < len(m.searchIdx) {
					idx = m.searchIdx[idx]
				}
				if idx < len(entries) {
					e := entries[idx]
					m.errMsg = ""
					m.queue.Clear()
					track := types.Track{
						Path:     e.Path,
						Title:    e.Title,
						Artist:   e.Artist,
						Album:    e.Album,
						CoverArt: e.CoverArt,
						Duration: e.Duration,
						Source:   types.Local,
					}
					m.queue.Add(track)
					m.queue.SetCurrent(0)
					m.queueCursor = 0
					m.queueOffset = 0
					return m.playTrack(track)
				}
			} else if m.active == "queue" && m.queue.Len() > 0 {
				idx := m.queueCursor
				if m.searchMode && idx < len(m.searchIdx) {
					idx = m.searchIdx[idx]
				}
				m.errMsg = ""
				m.queue.SetCurrent(idx)
				vis := m.queueVis()
				if m.queueCursor < m.queueOffset {
					m.queueOffset = m.queueCursor
				} else if m.queueCursor >= m.queueOffset+vis {
					m.queueOffset = m.queueCursor - vis + 1
				}
				tracks := m.queue.List()
				track := tracks[idx]
				return m.playTrack(track)
			}
		case " ", "space":
			if m.controller.State() == playback.Playing {
				m.controller.Pause()
			} else if m.controller.State() == playback.Paused {
				m.controller.Resume()
			}
		case "n":
			if m.queue.Len() > 0 {
				m.queueCursor = (m.queueCursor + 1) % m.queue.Len()
				m.queue.SetCurrent(m.queueCursor)
				vis := m.queueVis()
				if m.queueCursor >= m.queueOffset+vis {
					m.queueOffset = m.queueCursor - vis + 1
				}
				tracks := m.queue.List()
				track := tracks[m.queueCursor]
				return m.playTrack(track)
			}
		case "p":
			if m.queue.Len() > 0 {
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
				return m.playTrack(track)
			}
		case "[":
			m.controller.SetVolume(-0.1)
		case "]":
			m.controller.SetVolume(0.1)
		case "m":
			if m.controller.IsMuted() {
				m.controller.Unmute()
			} else {
				m.controller.Mute()
			}
		case "/":
			if !m.searchMode {
				m.searchMode = true
				m.searchQuery = ""
				if m.active == "library" {
					m.libOffset = 0
				} else if m.active == "tracks" {
					m.tracksOffset = 0
				} else {
					m.queueOffset = 0
				}
				return m, nil
			}
		case "<":
			if m.controller.State() != playback.Stopped {
				pos := m.controller.Position()
				dur := m.controller.Duration()
				newPos := pos - 5*time.Second
				if newPos < 0 {
					newPos = dur
				}
				m.controller.Seek(newPos)
			}
		case ">":
			if m.controller.State() != playback.Stopped {
				pos := m.controller.Position()
				dur := m.controller.Duration()
				newPos := pos + 5*time.Second
				if newPos > dur {
					newPos = dur
				}
				m.controller.Seek(newPos)
			}
		case "?":
			m.showHelp = !m.showHelp

		case "ctrl+s":
			m.sourceSwitch = true

		}

	case tickMsg:
		// Poll the Spotify device at most every 5s and only while a track is
		// actually playing. Polling every second (and polling forever while
		// idle) exhausted Spotify's playback quota -> 429 -> tracks won't play.
		cmds := []tea.Cmd{tick()}
		if m.controller.CurrentIsSpotify() && m.spotifyClient != nil && m.controller.State() == playback.Playing {
			if m.spotPollCount <= 0 {
				m.spotPollCount = 4
				cmds = append(cmds, spotifyPollState(m.spotifyClient))
			} else {
				m.spotPollCount--
			}
		}
		return m, tea.Batch(cmds...)

	case spotifyPollMsg:
		if !m.controller.CurrentIsSpotify() {
			return m, nil
		}
		// A nil state means 204 No Content: nothing is playing on the device,
		// which is what a finished single-track context looks like.
		ended := msg.state == nil
		if msg.state != nil {
			pos := time.Duration(msg.state.ProgressMs) * time.Millisecond
			dur := time.Duration(0)
			if msg.state.Item != nil {
				dur = time.Duration(msg.state.Item.DurationMs) * time.Millisecond
			}
			m.controller.UpdateSpotifySnapshot(pos, dur, msg.state.IsPlaying)
			ended = !msg.state.IsPlaying
		}
		// The device stopped on its own (track finished) while we expected
		// playback -> advance to the next queue track. A grace period avoids
		// advancing while the device is still starting the track.
		if ended && m.controller.State() == playback.Playing && m.controller.SpotGracePassed(3*time.Second) {
			cmd := m.advanceToNext()
			if cmd == nil {
				// Nothing left in the queue: go idle so polling stops
				// instead of hammering the API forever.
				m.controller.Stop()
				return m, nil
			}
			return m, cmd
		}
		return m, nil

	case songEndedMsg:
		if msg.id != m.playingID {
			return m, nil
		}
		if m.queue.Len() == 0 {
			return m, nil
		}
		return m, m.advanceToNext()

	case spotifyTracksMsg:
		m.spotifyLoading = false
		if msg.category != m.spotifyCategory || m.spotifyDrillType != "" {
			// stale hover response — a newer category is active
			return m, nil
		}
		if msg.offset == 0 {
			m.spotifyItems = msg.items
		} else {
			m.spotifyItems = append(m.spotifyItems, msg.items...)
		}
		m.spotifyTotal = msg.total
		m.spotifyOffset = msg.offset + len(msg.items)
		m.spotifyHasMore = m.spotifyOffset < msg.total
		return m, nil

	case spotifyPlaylistsMsg:
		m.spotifyLoading = false
		if msg.category != m.spotifyCategory || m.spotifyDrillType != "" {
			return m, nil
		}
		if msg.offset == 0 {
			m.spotifyPlaylists = msg.playlists
		} else {
			m.spotifyPlaylists = append(m.spotifyPlaylists, msg.playlists...)
		}
		m.spotifyTotal = msg.total
		m.spotifyListTotal = msg.total
		m.spotifyOffset = msg.offset + len(msg.playlists)
		m.spotifyHasMore = m.spotifyOffset < msg.total
		return m, nil

	case spotifyPlaylistTracksMsg:
		m.spotifyLoading = false
		if m.spotifyDrillType != "playlist" {
			// stale drill response
			return m, nil
		}
		if msg.offset == 0 {
			m.spotifyItems = msg.items
			m.spotifyDrillName = msg.name // name of the playlist
		} else {
			m.spotifyItems = append(m.spotifyItems, msg.items...)
		}
		m.spotifyTotal = msg.total
		m.spotifyOffset = msg.offset + len(msg.items)
		m.spotifyHasMore = m.spotifyOffset < msg.total
		return m, nil

	case spotifyAlbumsMsg:
		m.spotifyLoading = false
		if msg.category != m.spotifyCategory || m.spotifyDrillType != "" {
			return m, nil
		}
		if msg.offset == 0 {
			m.spotifyAlbums = msg.albums
		} else {
			m.spotifyAlbums = append(m.spotifyAlbums, msg.albums...)
		}
		m.spotifyTotal = msg.total
		m.spotifyListTotal = msg.total
		m.spotifyOffset = msg.offset + len(msg.albums)
		m.spotifyHasMore = m.spotifyOffset < msg.total
		return m, nil

	case spotifyAlbumTracksMsg:
		m.spotifyLoading = false
		if m.spotifyDrillType != "album" {
			return m, nil
		}
		if msg.offset == 0 {
			m.spotifyItems = msg.items
			m.spotifyDrillName = msg.name // name of the album
		} else {
			m.spotifyItems = append(m.spotifyItems, msg.items...)
		}
		m.spotifyTotal = msg.total
		m.spotifyOffset = msg.offset + len(msg.items)
		m.spotifyHasMore = m.spotifyOffset < msg.total
		return m, nil

	case spotifyErrorMsg:
		m.spotifyLoading = false
		m.spotifyErr = msg.err
		return m, nil

	}
	return m, nil
}

func (m model) handleSetupKey(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch m.setupState {
	case setupWelcome:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, m.quit()
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
	var src []string
	if m.active == "library" {
		for _, e := range m.browserDirs() {
			src = append(src, e.Name)
		}
	} else if m.active == "tracks" {
		for _, e := range m.trackFileEntries() {
			name := e.Title
			if name == "" {
				name = e.Name
			}
			src = append(src, name)
		}
	} else {
		for _, t := range m.queue.List() {
			name := t.Title
			if name == "" {
				name = t.Path
				if slash := strings.LastIndexByte(t.Path, '/'); slash >= 0 {
					name = t.Path[slash+1:]
				}
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

	mainH := m.height - 12
	libH := mainH * 30 / 100
	if libH < 3 {
		libH = 3
	}
	libVis := libH - 2
	if libVis < 0 {
		libVis = 0
	}
	queueH := mainH - libH
	queueVis := queueH - 2
	if queueVis < 0 {
		queueVis = 0
	}
	trackVis := m.height - 14
	if trackVis < 0 {
		trackVis = 0
	}

	if m.active == "library" {
		if m.libCursor >= len(m.searchIdx) {
			m.libCursor = len(m.searchIdx) - 1
		}
		if m.libOffset > m.libCursor {
			m.libOffset = m.libCursor
		}
		if libVis > 0 && m.libCursor >= m.libOffset+libVis {
			m.libOffset = m.libCursor - libVis + 1
		}
	}
	if m.active == "tracks" {
		if m.tracksCursor >= len(m.searchIdx) {
			m.tracksCursor = len(m.searchIdx) - 1
		}
		if m.tracksOffset > m.tracksCursor {
			m.tracksOffset = m.tracksCursor
		}
		if trackVis > 0 && m.tracksCursor >= m.tracksOffset+trackVis {
			m.tracksOffset = m.tracksCursor - trackVis + 1
		}
	}
	if m.active == "queue" {
		if m.queueCursor >= len(m.searchIdx) {
			m.queueCursor = len(m.searchIdx) - 1
		}
		if m.queueOffset > m.queueCursor {
			m.queueOffset = m.queueCursor
		}
		if queueVis > 0 && m.queueCursor >= m.queueOffset+queueVis {
			m.queueOffset = m.queueCursor - queueVis + 1
		}
	}
}

func (m model) buildBreadCrumb() string {
	parts := []string{"Library"}
	for _, h := range m.browser.History {
		parts = append(parts, filepath.Base(h))
	}
	parts = append(parts, m.browser.CurrentName())
	return strings.Join(parts, " > ")
}

func waitForSongEnd(player *playback.Player, id int) tea.Cmd {
	return func() tea.Msg {
		player.Wait()
		return songEndedMsg{id: id}
	}
}

func spotifyPollState(cl *spotify.Client) tea.Cmd {
	return func() tea.Msg {
		state, err := cl.GetPlaybackState()
		if err != nil {
			return nil
		}
		return spotifyPollMsg{state: state}
	}
}

// playbackCmds returns the tea.Cmd batch for whatever the controller just
// started playing: Spotify tracks are polled (there is no local stream to
// wait on), local tracks block on the beep player finishing.
func (m model) playbackCmds() tea.Cmd {
	var cmds []tea.Cmd
	if m.controller.CurrentIsSpotify() {
		cmds = append(cmds, tick())
	} else {
		cmds = append(cmds, waitForSongEnd(m.player, m.playingID), tick())
	}
	return tea.Batch(cmds...)
}

// playTrack starts playback of a track through the controller.
func (m model) playTrack(track types.Track) (tea.Model, tea.Cmd) {
	m.playingID++
	m.errMsg = ""
	if track.Source == types.Spotify {
		m.spotPollCount = 0
	}
	if err := m.controller.PlayTrack(track); err != nil {
		m.errMsg = fmt.Sprintf("Error: %v", err)
		return m, nil
	}
	return m, m.playbackCmds()
}

// advanceToNext moves the queue to the next track and starts it.
func (m model) advanceToNext() tea.Cmd {
	if m.queue.Len() == 0 {
		return nil
	}
	m.queueCursor = (m.queueCursor + 1) % m.queue.Len()
	m.queue.SetCurrent(m.queueCursor)
	vis := m.queueVis()
	if m.queueCursor >= m.queueOffset+vis {
		m.queueOffset = m.queueCursor - vis + 1
	}

	tracks := m.queue.List()
	track := tracks[m.queueCursor]
	m.playingID++
	m.errMsg = ""
	if track.Source == types.Spotify {
		m.spotPollCount = 0
	}
	if err := m.controller.PlayTrack(track); err != nil {
		m.errMsg = fmt.Sprintf("Error: %v", err)
		return nil
	}
	return m.playbackCmds()
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) quit() tea.Cmd {
	if m.controller != nil {
		m.controller.PauseSpotify()
	}
	return tea.Quit
}

func (m model) setupView() tea.View {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#000000"))
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
	s += titleStyle.Render("\x1b[48;2;180;190;254m Welcome to Resonance ") + "\n\n"
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
		s += keyStyle.Render("  " + common.Confirm + " confirm    " + "<esc>" + " cancel")
	}

	box := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s)
	return tea.View{
		AltScreen: true,
		Content:   m.fillBg(box),
	}
}

func splitAtVisualPos(s string, pos int) (string, string) {
	var before, rest strings.Builder
	visualCol := 0
	runes := []rune(s)
	i := 0
	for i < len(runes) {
		if runes[i] == '\x1b' {
			before.WriteRune(runes[i])
			i++
			for i < len(runes) {
				before.WriteRune(runes[i])
				ch := runes[i]
				if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
					i++
					break
				}
				i++
			}
			continue
		}
		if visualCol < pos {
			before.WriteRune(runes[i])
		} else {
			rest.WriteRune(runes[i])
		}
		visualCol++
		i++
	}
	return before.String(), rest.String()
}

func placeOverlay(termW, termH int, overlay, bg string) string {
	bgLines := strings.Split(bg, "\n")
	oLines := strings.Split(overlay, "\n")

	boxW := 0
	for _, l := range oLines {
		if w := lipgloss.Width(l); w > boxW {
			boxW = w
		}
	}
	boxH := len(oLines)

	ox := (termW - boxW) / 2
	oy := (termH - boxH) / 2

	for len(bgLines) < termH {
		bgLines = append(bgLines, strings.Repeat(" ", termW))
	}

	for i, oLine := range oLines {
		row := oy + i
		if row < 0 || row >= len(bgLines) {
			continue
		}
		left, right := splitAtVisualPos(bgLines[row], ox)
		_, rightAfterBox := splitAtVisualPos(right, boxW)
		bgLines[row] = left + oLine + rightAfterBox
	}

	return strings.Join(bgLines, "\n")
}

func (m model) tooSmallView() tea.View {
	badStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#d46159")).Bold(true)

	wStr := fmt.Sprintf("%d", m.width)
	if m.width < 84 {
		wStr = badStyle.Render(fmt.Sprintf("%d", m.width))
	}
	hStr := fmt.Sprintf("%d", m.height)
	if m.height < 28 {
		hStr = badStyle.Render(fmt.Sprintf("%d", m.height))
	}

	s := fmt.Sprintf("Terminal size too small\nWidth = %s  Height = %s\n\nNeeded for current config\nWidth >= 84, Height >= 28", wStr, hStr)
	placed := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s)
	return tea.View{
		AltScreen: true,
		Content:   m.fillBg(placed),
	}
}

func (m model) View() tea.View {
	if m.setup {
		return m.setupView()
	}
	if m.tooSmall {
		return m.tooSmallView()
	}

	s := m.buildMainView()

	if m.sourceSwitch {
		m.sourceSwitchPanel.SetData(components.SourceSwitchData{
			Visible:      true,
			SourceCursor: m.sourceCursor,
			SpotifyOK:    m.spotifyLoggedIn,
			Width:        m.width,
			Height:       m.height,
		})
		box := m.sourceSwitchPanel.View()
		s = placeOverlay(m.width, m.height, box, s)
	}

	if m.showHelp {
		box := m.helpBox()
		s = placeOverlay(m.width, m.height, box, s)
	}

	return tea.View{
		AltScreen: true,
		Content:   m.fillBg(s),
	}
}

// func dummyEnteries() []types.Entry {
// 	var dir []types.Entry
// 	for i := 0; i < 5; i++ {
// 		dir = append(dir,
// 			types.Entry{
// 				Name:     fmt.Sprintf("hello%d", i),
// 				Path:     fmt.Sprintf("path%d", i),
// 				IsDir:    false,
// 				Title:    fmt.Sprintf("title%d", i),
// 				Artist:   fmt.Sprintf("artist%d", i),
// 				Album:    fmt.Sprintf("album%d", i),
// 				Duration: time.Duration(0), // Or omit entirely for default 0 value
// 				CoverArt: []byte{},         // Or nil
// 			},
// 		)
// 	}
// 	return dir
// }

func (m model) buildMainView() string {
	topHeight := 1
	footerHeight := 11
	mainHeight := m.contentHeight()

	leftWidth := m.width * 35 / 100
	rightWidth := m.width - leftWidth

	libHeight := m.libHeight()
	queueHeight := mainHeight - libHeight

	// filter dir only entries for library
	if m.source == "spotify" {
		m.libraryPanel.SetData(components.LibData{
			Entries:     spotifyCategories(),
			Title:       "Spotify",
			Breadcrumb:  "Spotify",
			Cursor:      m.spotifyCursor,
			Offset:      0,
			Total:       3,
			Active:      m.active == "library",
			Width:       leftWidth,
			Height:      libHeight,
			SearchMode:  false,
			SearchQuery: "",
		})
	} else {
		dirs := m.browserDirs()
		if m.active == "library" && m.searchMode && len(m.searchIdx) > 0 {
			filtered := make([]types.Entry, len(m.searchIdx))
			for i, idx := range m.searchIdx {
				filtered[i] = dirs[idx]
			}
			dirs = filtered
		}
		dirSlice := dirs
		dirOffset := 0
		if len(dirs) > libHeight-2 {
			off := m.libOffset
			if off > len(dirs)-libHeight+2 {
				off = len(dirs) - libHeight + 2
			}
			if off < 0 {
				off = 0
			}
			dirSlice = dirs[off : off+libHeight-2]
			dirOffset = off
		}

		m.libraryPanel.SetData(components.LibData{
			Entries:     dirSlice,
			Title:       m.browser.CurrentName(),
			Breadcrumb:  m.buildBreadCrumb(),
			Cursor:      m.libCursor,
			Offset:      dirOffset,
			Total:       len(dirs),
			Active:      m.active == "library",
			Width:       leftWidth,
			Height:      libHeight,
			SearchMode:  m.searchMode && m.active == "library",
			SearchQuery: m.searchQuery,
		})
	}

	// queue panel (below library)
	tracks := m.queue.List()
	if m.active == "queue" && m.searchMode && len(m.searchIdx) > 0 {
		filtered := make([]types.Track, len(m.searchIdx))
		for i, idx := range m.searchIdx {
			filtered[i] = tracks[idx]
		}
		tracks = filtered
	}
	queueSlice := tracks
	queueOffset := 0
	queueVis := m.queueVis()
	if len(tracks) > queueVis {
		off := m.queueOffset
		if off > len(tracks)-queueVis {
			off = len(tracks) - queueVis
		}
		if off < 0 {
			off = 0
		}
		queueSlice = tracks[off : off+queueVis]
		queueOffset = off
	}
	// PlayingIdx relative to filtered list
	playIdx := m.queue.CurrentIndex()
	if m.active == "queue" && m.searchMode && len(m.searchIdx) > 0 {
		playIdx = -1
		for i, idx := range m.searchIdx {
			if idx == m.queue.CurrentIndex() {
				playIdx = i
				break
			}
		}
	}
	m.queuePanel.SetData(components.QueueData{
		Tracks:      queueSlice,
		PlayingIdx:  playIdx,
		Playing:     m.controller.State() != playback.Stopped,
		Cursor:      m.queueCursor,
		Offset:      queueOffset,
		Total:       len(tracks),
		Active:      m.active == "queue",
		Width:       leftWidth,
		Height:      queueHeight,
		SearchMode:  m.searchMode && m.active == "queue",
		SearchQuery: m.searchQuery,
	})

	// tracks panel (right side)
	playingTrack := m.controller.CurrentTrack()

	if m.source == "spotify" {
		if m.spotifyErr != "" {
			m.tracksPanel.SetData(components.TracksData{
				ContentType:  "error",
				ErrorMessage: m.spotifyErr,
				Title:        "Spotify",
				Active:       m.active == "tracks",
				Height:       mainHeight,
				Width:        rightWidth,
			})
		} else {
			trackVis := m.trackVis()

			var contentType, title string
			var trackSlice []types.Track
			var playlistSlice []spotify.SpotifyPlaylist
			var albumSlice []spotify.SavedAlbum
			switch {
			case m.spotifyDrillType != "":
				contentType, title = "spotify-tracks", m.spotifyDrillName
			case m.spotifyCategory == 0:
				contentType, title = "spotify-tracks", "Liked Songs"
			case m.spotifyCategory == 1:
				contentType, title = "spotify-playlists", "Playlists"
			case m.spotifyCategory == 2:
				contentType, title = "spotify-albums", "Albums"
			}

			activeListLen := 0
			switch contentType {
			case "spotify-playlists":
				activeListLen = len(m.spotifyPlaylists)
			case "spotify-albums":
				activeListLen = len(m.spotifyAlbums)
			default:
				activeListLen = len(m.spotifyItems)
			}

			contentOffset := 0
			if activeListLen > trackVis {
				off := m.spotifyScroll
				if off > activeListLen-trackVis {
					off = activeListLen - trackVis
				}
				if off < 0 {
					off = 0
				}
				if contentType == "spotify-playlists" {
					playlistSlice = m.spotifyPlaylists[off : off+trackVis]
				} else if contentType == "spotify-albums" {
					albumSlice = m.spotifyAlbums[off : off+trackVis]
				} else {
					trackSlice = m.spotifyItems[off : off+trackVis]
				}
				contentOffset = off
			} else {
				switch contentType {
				case "spotify-playlists":
					playlistSlice = m.spotifyPlaylists
				case "spotify-albums":
					albumSlice = m.spotifyAlbums
				default:
					trackSlice = m.spotifyItems
				}
			}

			m.tracksPanel.SetData(components.TracksData{
				Entries:          nil,
				SpotifyTracks:    trackSlice,
				SpotifyPlaylists: playlistSlice,
				SpotifyAlbums:    albumSlice,
				HideAlbum:        m.spotifyDrillType == "album",
				ContentType:      contentType,
				Title:            title,
				PlayingPath:      "",
				Cursor:           m.spotifyContentCursor,
				Offset:           contentOffset,
				Total:            m.spotifyTotal,
				Active:           m.active == "tracks",
				Height:           mainHeight,
				Width:            rightWidth,
				SearchMode:       false,
				Loading:          m.spotifyLoading,
			})
		}
	} else {
		fileEntries := m.fileEntries
		if fileEntries == nil {
			fileEntries = m.trackFileEntries()
		}
		if m.active == "tracks" && m.searchMode && len(m.searchIdx) > 0 {
			filtered := make([]types.Entry, len(m.searchIdx))
			for i, idx := range m.searchIdx {
				filtered[i] = fileEntries[idx]
			}
			fileEntries = filtered
		}
		fileSlice := fileEntries
		fileOffset := 0
		trackVis := m.trackVis()
		if len(fileEntries) > trackVis {
			off := m.tracksOffset
			if off > len(fileEntries)-trackVis {
				off = len(fileEntries) - trackVis
			}
			if off < 0 {
				off = 0
			}
			fileSlice = fileEntries[off : off+trackVis]
			fileOffset = off
		}
		m.tracksPanel.SetData(components.TracksData{
			Entries:     fileSlice,
			PlayingPath: playingTrack.Path,
			Cursor:      m.tracksCursor,
			Offset:      fileOffset,
			Total:       len(fileEntries),
			Active:      m.active == "tracks",
			Height:      mainHeight,
			Width:       rightWidth,
			SearchMode:  m.searchMode && m.active == "tracks",
			SearchQuery: m.searchQuery,
			SubdirCount: m.subdirCount,
		})
	}

	trackName := playingTrack.Title
	if trackName == "" {
		trackName = playingTrack.Path
		if idx := strings.LastIndexByte(playingTrack.Path, '/'); idx >= 0 {
			trackName = playingTrack.Path[idx+1:]
		}
	}

	title := playingTrack.Title
	if title == "" {
		title = "Unknown Title"
	}
	album := playingTrack.Album
	if album == "" {
		album = "Unknown Album"
	}
	artist := playingTrack.Artist
	if artist == "" {
		artist = "Unknown Artist"
	}

	var stateIcon string
	switch m.controller.State() {
	case playback.Playing:
		stateIcon = pausedIconStyle.Render(common.Pause)
	case playback.Paused:
		stateIcon = playingIconStyle.Render(common.Play)
	default:
		stateIcon = common.MusicNote2
	}

	m.footer.SetData(components.FooterData{
		TrackName: trackName,
		Title:     title,
		Album:     album,
		Artist:    artist,
		StateIcon: stateIcon,
		Position:  m.controller.Position(),
		Duration:  m.controller.Duration(),
		VolLevel:  m.controller.Volume(),
		Muted:     m.controller.IsMuted(),
		Height:    footerHeight,
		Width:     m.width,
		CoverArt:  playingTrack.CoverArt,
	})

	m.top.SetTrackCount(m.browser.TrackCount(), m.source)

	leftCol := lipgloss.JoinVertical(lipgloss.Left,
		m.libraryPanel.View(),
		m.queuePanel.View(),
	)

	s := lipgloss.JoinVertical(
		lipgloss.Left,
		m.top.View(topHeight, m.width),
		lipgloss.JoinHorizontal(lipgloss.Top, leftCol, m.tracksPanel.View()),
		m.footer.View(),
	)

	return s
}

func (m model) helpBox() string {
	helpW := 54
	header := lipgloss.NewStyle().Bold(true).Italic(true).Foreground(lipgloss.Color("#FFFFFF"))
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#B4BEFE"))
	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render("│")
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	subheader := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A6E3A1"))

	groups := []struct {
		section string
		items   [][2]string
	}{
		{"Navigation", [][2]string{
			{"↑ / k , ↓ / j", "Move cursor (preview "},
			{"", "content)"},
			{"Tab", "Switch panel"},
			{"Enter", "Drill into directory"},
			{"Backspace / h", "Parent directory"},
		}},
		{"Search", [][2]string{
			{"/", "Enter search mode"},
			{"Esc", "Exit search"},
			{"Backspace", "Delete character"},
		}},
		{"Add to Queue", [][2]string{
			{"a (track/song)", "Add single to queue"},
			{"a (dir)", "Add all from dir"},
			{"A", "Add all from selected dir"},
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
			{"esc", "Close"},
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

	hstr := header.AlignHorizontal(lipgloss.Center).AlignVertical(lipgloss.Center).Render("Keyboard Shortcuts")

	var bar string
	for i := 0; i < helpW-9; i++ {
		// bar += "─"
		bar += " "
	}

	b.WriteString(hstr + "\n")
	b.WriteString(bar + "\n")

	for _, g := range groups {
		b.WriteString(subheader.Render(g.section) + "\n")
		for _, item := range g.items {
			key := item[0]
			paddingLen := maxKeyW - lipgloss.Width(key)
			if paddingLen < 0 {
				paddingLen = 0
			}
			padded := key + strings.Repeat(" ", paddingLen)

			line := fmt.Sprintf("%s %s %s", keyStyle.Render(padded), sep, item[1])
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(dim.Render("  ? or Esc to close"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#A6E3A1")).
		Padding(1, 3).
		Width(helpW).
		Render(b.String())

	return box
}
