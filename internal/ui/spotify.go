// Ui code that calls the spotify package
package tui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/yadneshx17/resonance/internal/spotify"
	"github.com/yadneshx17/resonance/internal/types"
)

func spotifyCategories() []types.Entry {
	return []types.Entry{
		{Name: "Liked Songs"},
		{Name: "Playlists"},
		{Name: "Albums"},
	}
}

// loadCategory resets all content-panel state and fetches the first page of
// the given category (0=Liked Songs, 1=Playlists, 2=Albums). Used by
// hover-to-load (M5): moving the cursor over a category loads its content.
func (m model) loadCategory(cat int) (model, tea.Cmd) {
	m.spotifyCategory = cat
	m.spotifyDrillType = ""
	m.spotifyItems = nil
	m.spotifyPlaylists = nil
	m.spotifyAlbums = nil
	m.spotifyTotal = 0
	m.spotifyOffset = 0
	m.spotifyListTotal = 0
	m.spotifyLoading = true
	m.spotifyErr = ""
	m.spotifyContentCursor = 0
	m.spotifyScroll = 0
	m.spotifyContentCursorBackup = 0
	m.spotifyScrollBackup = 0
	switch cat {
	case 0:
		return m, m.fetchSpotifyLiked(0)
	case 1:
		return m, m.fetchSpotifyPlaylists(0)
	case 2:
		return m, m.fetchSpotifyAlbums(0)
	}
	return m, nil
}

// spotifyContentLen returns the number of items currently loaded in the
// content panel for whatever view is active.
func (m model) spotifyContentLen() int {
	if m.spotifyDrillType != "" || m.spotifyCategory == 0 {
		return len(m.spotifyItems)
	}
	if m.spotifyCategory == 1 {
		return len(m.spotifyPlaylists)
	}
	if m.spotifyCategory == 2 {
		return len(m.spotifyAlbums)
	}
	return 0
}

func (m model) handleSpotifyUp() (model, tea.Cmd) {
	if m.active == "library" {
		if m.spotifyCursor > 0 {
			m.spotifyCursor--
			return m.loadCategory(m.spotifyCursor)
		}
		return m, nil
	}
	if m.active == "tracks" {
		if m.spotifyContentCursor > 0 {
			m.spotifyContentCursor--
		}
		if m.spotifyContentCursor < m.spotifyScroll {
			m.spotifyScroll = m.spotifyContentCursor
		}
		return m, nil
	}
	return m, nil
}

func (m model) handleSpotifyDown() (model, tea.Cmd) {
	if m.active == "library" {
		maxLen := len(spotifyCategories())
		if m.spotifyCursor < maxLen-1 {
			m.spotifyCursor++
			return m.loadCategory(m.spotifyCursor)
		}
		return m, nil
	}
	if m.active == "tracks" {
		maxLen := m.spotifyTotal
		if m.spotifyContentCursor < maxLen-1 {
			m.spotifyContentCursor++
		}
		trackVis := m.trackVis()
		if m.spotifyContentCursor >= m.spotifyScroll+trackVis {
			m.spotifyScroll = m.spotifyContentCursor - trackVis + 1
		}
		return m, m.checkInfiniteScroll()
	}
	return m, nil
}

// handleSpotifyContentEnter handles Enter in the content panel (M7):
//   - showing a playlist/album list  → drill into the highlighted item
//   - showing tracks (liked/drill)   → add the highlighted track to the queue
func (m model) handleSpotifyContentEnter() (model, tea.Cmd) {
	switch m.spotifyDrillType {
	case "playlist", "album":
		if m.spotifyContentCursor < len(m.spotifyItems) && m.spotifyItems[m.spotifyContentCursor].ID != "" {
			m.queue.Add(m.spotifyItems[m.spotifyContentCursor])
		}
		return m, nil
	default:
		switch m.spotifyCategory {
		case 1: // playlists list → drill
			if m.spotifyContentCursor < len(m.spotifyPlaylists) {
				pl := m.spotifyPlaylists[m.spotifyContentCursor]
				m.spotifyDrillType = "playlist"
				m.spotifyDrillID = pl.ID
				m.spotifyDrillName = pl.Name
				m.spotifyItems = nil
				m.spotifyTotal = 0
				m.spotifyOffset = 0
				m.spotifyLoading = true
				m.spotifyErr = ""
				m.spotifyContentCursorBackup = m.spotifyContentCursor
				m.spotifyScrollBackup = m.spotifyScroll
				m.spotifyContentCursor = 0
				m.spotifyScroll = 0
				return m, m.fetchSpotifyPlaylistTracks(pl.ID, 0)
			}
		case 2: // albums list → drill
			if m.spotifyContentCursor < len(m.spotifyAlbums) {
				al := m.spotifyAlbums[m.spotifyContentCursor].Album
				m.spotifyDrillType = "album"
				m.spotifyDrillID = al.ID
				m.spotifyDrillName = al.Name
				m.spotifyItems = nil
				m.spotifyTotal = 0
				m.spotifyOffset = 0
				m.spotifyLoading = true
				m.spotifyErr = ""
				m.spotifyContentCursorBackup = m.spotifyContentCursor
				m.spotifyScrollBackup = m.spotifyScroll
				m.spotifyContentCursor = 0
				m.spotifyScroll = 0
				return m, m.fetchSpotifyAlbumTracks(al.ID, 0)
			}
		case 0: // Liked Songs tracks → add to queue
			if m.spotifyContentCursor < len(m.spotifyItems) && m.spotifyItems[m.spotifyContentCursor].ID != "" {
				m.queue.Add(m.spotifyItems[m.spotifyContentCursor])
			}
		}
	}
	return m, nil
}

func (m model) handleSpotifyBackspace() (model, tea.Cmd) {
	if m.active == "tracks" && m.spotifyDrillType != "" {
		// Exit drill back to the category list. No refetch needed — the
		// playlists/albums lists are still in memory. Restore the position
		// saved when the drill was entered (N2): the content cursor only
		// resets when the category changes, not when drilling in/out.
		m.spotifyDrillType = ""
		m.spotifyItems = nil
		m.spotifyContentCursor = m.spotifyContentCursorBackup
		m.spotifyScroll = m.spotifyScrollBackup
		m.spotifyErr = ""
		// Restore the list's pagination state (drill fetch overwrote it).
		// spotifyContentLen() is called after clearing drillType so it
		// reflects the playlists/albums list length.
		loaded := m.spotifyContentLen()
		m.spotifyTotal = m.spotifyListTotal
		m.spotifyOffset = loaded
		m.spotifyHasMore = loaded < m.spotifyListTotal
		// Clamp the restored cursor/scroll so the cursor stays visible.
		if loaded > 0 && m.spotifyContentCursor >= loaded {
			m.spotifyContentCursor = loaded - 1
		}
		vis := m.trackVis()
		if vis > 0 && m.spotifyContentCursor >= m.spotifyScroll+vis {
			m.spotifyScroll = m.spotifyContentCursor - vis + 1
		}
		if m.spotifyScroll < 0 {
			m.spotifyScroll = 0
		}
		return m, nil
	}
	return m, nil
}

func (m model) handleSpotifySmallA() (model, tea.Cmd) {
	if m.active == "tracks" && m.spotifyContentCursor < len(m.spotifyItems) && m.spotifyItems[m.spotifyContentCursor].ID != "" {
		m.queue.Add(m.spotifyItems[m.spotifyContentCursor])
	}
	return m, nil
}

// checkInfiniteScroll fetches the next page when the content cursor nears the
// end of the loaded list (M8). Branches by drill state / category.
func (m model) checkInfiniteScroll() tea.Cmd {
	if m.source != "spotify" || !m.spotifyHasMore || m.spotifyLoading {
		return nil
	}
	loaded := m.spotifyContentLen()
	if m.spotifyContentCursor >= loaded-3 {
		if m.spotifyDrillType == "playlist" {
			return m.fetchSpotifyPlaylistTracks(m.spotifyDrillID, m.spotifyOffset)
		}
		if m.spotifyDrillType == "album" {
			return m.fetchSpotifyAlbumTracks(m.spotifyDrillID, m.spotifyOffset)
		}
		switch m.spotifyCategory {
		case 0:
			return m.fetchSpotifyLiked(m.spotifyOffset)
		case 1:
			return m.fetchSpotifyPlaylists(m.spotifyOffset)
		case 2:
			return m.fetchSpotifyAlbums(m.spotifyOffset)
		}
	}
	return nil
}

// tea.Cmd fetch functions
func (m model) fetchSpotifyLiked(offset int) tea.Cmd {
	return func() tea.Msg {
		result, err := m.spotifyClient.GetSavedTracks(50, offset)
		if err != nil {
			return spotifyErrorMsg{err: err.Error()}
		}
		items := make([]types.Track, 0, len(result.Items))
		for _, st := range result.Items {
			items = append(items, spotifyTrackToTrack(st.Track))
		}
		return spotifyTracksMsg{
			items:    items,
			total:    result.Total,
			offset:   offset,
			category: 0,
		}
	}
}

func (m model) fetchSpotifyPlaylists(offset int) tea.Cmd {
	return func() tea.Msg {
		result, err := m.spotifyClient.GetPlaylists(50, offset)
		if err != nil {
			return spotifyErrorMsg{err: err.Error()}
		}
		return spotifyPlaylistsMsg{
			playlists: result.Items,
			total:     result.Total,
			offset:    offset,
			category:  1,
		}
	}
}

func (m model) fetchSpotifyPlaylistTracks(id string, offset int) tea.Cmd {
	return func() tea.Msg {
		result, err := m.spotifyClient.GetPlaylistTracks(id, 50, offset)
		if err != nil {
			return spotifyErrorMsg{err: friendlyPlaylistErr(err)}
		}
		items := make([]types.Track, 0, len(result.Items))
		for _, item := range result.Items {
			items = append(items, spotifyTrackToTrack(item.Track))
		}
		return spotifyPlaylistTracksMsg{
			items:  items,
			total:  result.Total,
			offset: offset,
			name:   m.spotifyDrillName,
		}
	}
}

// friendlyPlaylistErr turns Spotify's opaque 403 into a message the user can
// act on. Since Feb 2026 Spotify only exposes a playlist's tracks to its
// owner or collaborators — even public playlists return 403 otherwise.
func friendlyPlaylistErr(err error) string {
	if strings.Contains(err.Error(), "(403)") {
		return "Playlist not accessible (403). Since Feb 2026 Spotify only lets you read playlists you own or collaborate on, even public ones. If it's yours, re-authorize the app: delete ~/.config/resonance/credentials/spotify.json and run setup again."
	}
	return err.Error()
}

func (m model) fetchSpotifyAlbums(offset int) tea.Cmd {
	return func() tea.Msg {
		result, err := m.spotifyClient.GetSavedAlbums(50, offset)
		if err != nil {
			return spotifyErrorMsg{err: err.Error()}
		}
		return spotifyAlbumsMsg{
			albums:   result.Items,
			total:    result.Total,
			offset:   offset,
			category: 2,
		}
	}
}

func (m model) fetchSpotifyAlbumTracks(id string, offset int) tea.Cmd {
	return func() tea.Msg {
		result, err := m.spotifyClient.GetAlbumTracks(id, 50, offset)
		if err != nil {
			return spotifyErrorMsg{err: err.Error()}
		}
		items := make([]types.Track, 0, len(result.Items))
		for _, t := range result.Items {
			items = append(items, spotifyTrackToTrack(t))
		}
		return spotifyAlbumTracksMsg{
			items:  items,
			total:  result.Total,
			offset: offset,
			name:   m.spotifyDrillName,
		}
	}
}

// Utility
func spotifyTrackToTrack(st spotify.SpotifyTrack) types.Track {
	artist := ""
	if len(st.Artists) > 0 {
		artist = st.Artists[0].Name
	}
	album := ""
	if st.Album.Name != "" {
		album = st.Album.Name
	}
	return types.Track{
		ID:       st.ID,
		Title:    st.Name,
		Artist:   artist,
		Album:    album,
		Source:   types.Spotify,
		Duration: time.Duration(st.DurationMs) * time.Millisecond,
	}
}
