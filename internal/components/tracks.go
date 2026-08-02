package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/yadneshx17/resonance/internal/common"
	"github.com/yadneshx17/resonance/internal/renderer"
	"github.com/yadneshx17/resonance/internal/spotify"
	"github.com/yadneshx17/resonance/internal/types"
)

type TracksData struct {
	Entries          []types.Entry
	PlayingPath      string
	Cursor           int
	Offset           int
	Total            int
	Active           bool
	Height           int
	Width            int
	SearchMode       bool
	SearchQuery      string
	SubdirCount      int
	Loading          bool
	ContentType      string                    // "local", "spotify-tracks", "spotify-playlists", "spotify-albums", "error"
	SpotifyTracks    []types.Track             // when ContentType == "spotify-tracks"
	SpotifyPlaylists []spotify.SpotifyPlaylist // when ContentType == "spotify-playlists"
	SpotifyAlbums    []spotify.SavedAlbum      // when ContentType == "spotify-albums"
	HideAlbum        bool                      // album drill: drop the Album column (every row is the same album)
	ErrorMessage     string                    // when ContentType == "error"
	Title            string                    // panel title (replaces hardcoded "Content")
}

type Tracks struct {
	entries          []types.Entry
	playingPath      string
	cursor           int
	offset           int
	total            int
	active           bool
	height           int
	width            int
	searchMode       bool
	searchQuery      string
	subdirCount      int
	loading          bool
	contentType      string
	spotifyTracks    []types.Track
	spotifyPlaylists []spotify.SpotifyPlaylist
	spotifyAlbums    []spotify.SavedAlbum
	hideAlbum        bool
	errorMessage     string
	title            string
}

func (t *Tracks) SetData(data TracksData) {
	t.entries = data.Entries
	t.playingPath = data.PlayingPath
	t.cursor = data.Cursor
	t.offset = data.Offset
	t.total = data.Total
	t.active = data.Active
	t.height = data.Height
	t.width = data.Width
	t.searchMode = data.SearchMode
	t.searchQuery = data.SearchQuery
	t.loading = data.Loading
	t.subdirCount = data.SubdirCount
	t.contentType = data.ContentType
	t.spotifyTracks = data.SpotifyTracks
	t.spotifyPlaylists = data.SpotifyPlaylists
	t.spotifyAlbums = data.SpotifyAlbums
	t.hideAlbum = data.HideAlbum
	t.errorMessage = data.ErrorMessage
	t.title = data.Title
}

func (t Tracks) buildBlock() string {
	if t.width < 4 {
		return ""
	}

	contentWidth := t.width - 4
	var lines []string
	if t.searchMode {
		lines = append(lines, common.SearchBar(t.searchQuery, contentWidth))
	}

	seqWidth := 3
	prefixWidth := 3
	durWidth := 6
	remaining := contentWidth - seqWidth - 1 - prefixWidth - 3 - durWidth
	titleWidth := remaining * 34 / 100
	artistWidth := remaining * 33 / 100
	albumWidth := remaining - titleWidth - artistWidth
	if titleWidth < 3 {
		titleWidth = 3
	}
	if artistWidth < 3 {
		artistWidth = 3
	}
	if albumWidth < 3 {
		albumWidth = 3
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#6C7086"))
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#45475A"))

	header := fmt.Sprintf("%*s %-*s%-*s %-*s %-*s %*s",
		seqWidth, "#",
		prefixWidth, "",
		titleWidth, "Title",
		artistWidth, "Artist",
		albumWidth, "Album",
		durWidth, "Dur",
	)
	header = headerStyle.Render(header)

	sep := fmt.Sprintf("%s %s%s %s %s %s",
		strings.Repeat("─", seqWidth),
		strings.Repeat("─", prefixWidth),
		strings.Repeat("─", titleWidth),
		strings.Repeat("─", artistWidth),
		strings.Repeat("─", albumWidth),
		strings.Repeat("─", durWidth),
	)
	sep = sepStyle.Render(sep)

	// Album drill: every row belongs to one album, so drop the Album column and
	// split its width between Title and Artist.
	if t.hideAlbum {
		extraTitle := albumWidth / 2
		artistWidth += albumWidth - extraTitle
		titleWidth += extraTitle
		header = headerStyle.Render(fmt.Sprintf("%*s %-*s%-*s %-*s %*s",
			seqWidth, "#",
			prefixWidth, "",
			titleWidth, "Title",
			artistWidth, "Artist",
			durWidth, "Dur",
		))
		sep = sepStyle.Render(fmt.Sprintf("%s %s%s %s %s",
			strings.Repeat("─", seqWidth),
			strings.Repeat("─", prefixWidth),
			strings.Repeat("─", titleWidth),
			strings.Repeat("─", artistWidth),
			strings.Repeat("─", durWidth),
		))
	}

	if t.contentType == "error" {
		style := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center)
		msg := t.errorMessage
		if msg == "" {
			msg = "Something went wrong"
		}
		return style.Render(msg)
	}

	if t.contentType == "spotify-playlists" {
		if t.loading && len(t.spotifyPlaylists) == 0 {
			style := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center)
			return style.Render("Loading...")
		}
		if len(t.spotifyPlaylists) == 0 {
			style := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center)
			return style.Render("No playlists")
		}
		songsW := 11
		visW := 8
		nameW := contentWidth - seqWidth - 1 - prefixWidth - 1 - songsW - 1 - visW
		if nameW < 3 {
			nameW = 3
		}
		lines = append(lines, headerStyle.Render(fmt.Sprintf("%*s %-*s%-*s %-*s %s",
			seqWidth, "#",
			prefixWidth, "",
			nameW, "Name",
			songsW, "Songs",
			"Vis",
		)))
		lines = append(lines, sepStyle.Render(fmt.Sprintf("%s %s%s %s %s",
			strings.Repeat("─", seqWidth),
			strings.Repeat("─", prefixWidth),
			strings.Repeat("─", nameW),
			strings.Repeat("─", songsW),
			strings.Repeat("─", visW),
		)))
		for i, pl := range t.spotifyPlaylists {
			idx := t.offset + i
			name := pl.Name
			vis := "—"
			if pl.Public != nil {
				if *pl.Public {
					vis = "Public"
				} else {
					vis = "Private"
				}
			}
			count := fmt.Sprintf("%d songs", pl.Tracks.Total)
			runes := []rune(name)
			if len(runes) > nameW {
				name = string(runes[:nameW-1]) + "…"
			}
			prefix := "   "
			if t.active && t.cursor == idx {
				prefix = common.Cursor
			}
			seq := fmt.Sprintf("%*d", seqWidth, idx+1)
			line := fmt.Sprintf("%s %s%-*s %-*s %s",
				seq, prefix,
				nameW, name,
				songsW, count,
				vis,
			)
			if t.active && t.cursor == idx {
				line = common.CursorStyle.Render(line)
			}
			lines = append(lines, line)
		}
		return strings.Join(lines, "\n")
	}

	if t.contentType == "spotify-albums" {
		if t.loading && len(t.spotifyAlbums) == 0 {
			style := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center)
			return style.Render("Loading...")
		}
		if len(t.spotifyAlbums) == 0 {
			style := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center)
			return style.Render("No albums")
		}
		songsW := 11
		nameW := contentWidth - seqWidth - 1 - prefixWidth - 1 - songsW
		if nameW < 3 {
			nameW = 3
		}
		lines = append(lines, headerStyle.Render(fmt.Sprintf("%*s %-*s%-*s %-*s",
			seqWidth, "#",
			prefixWidth, "",
			nameW, "Name",
			songsW, "Songs",
		)))
		lines = append(lines, sepStyle.Render(fmt.Sprintf("%s %s%s %s",
			strings.Repeat("─", seqWidth),
			strings.Repeat("─", prefixWidth),
			strings.Repeat("─", nameW),
			strings.Repeat("─", songsW),
		)))
		for i, sa := range t.spotifyAlbums {
			idx := t.offset + i
			name := sa.Album.Name
			count := fmt.Sprintf("%d songs", sa.Album.TotalTracks)
			runes := []rune(name)
			if len(runes) > nameW {
				name = string(runes[:nameW-1]) + "…"
			}
			prefix := "   "
			if t.active && t.cursor == idx {
				prefix = common.Cursor
			}
			seq := fmt.Sprintf("%*d", seqWidth, idx+1)
			line := fmt.Sprintf("%s %s%-*s %-*s",
				seq, prefix,
				nameW, name,
				songsW, count,
			)
			if t.active && t.cursor == idx {
				line = common.CursorStyle.Render(line)
			}
			lines = append(lines, line)
		}
		return strings.Join(lines, "\n")
	}

	if t.contentType == "spotify-tracks" {
		if t.loading && len(t.spotifyTracks) == 0 {
			style := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center)
			return style.Render("Loading...")
		}
		if len(t.spotifyTracks) == 0 {
			style := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center)
			return style.Render("No tracks in this library")
		}
		lines = append(lines, header, sep)
		for i, tr := range t.spotifyTracks {
			idx := t.offset + i
			title := tr.Title
			artist := tr.Artist
			album := tr.Album
			if tr.ID == "" {
				// Spotify returns null track objects for removed/unavailable
				// entries — show a placeholder so row numbering stays aligned.
				title = "(Unavailable)"
				artist = "—"
				album = "—"
			}

			runesT := []rune(title)
			if len(runesT) > titleWidth {
				title = string(runesT[:titleWidth-1]) + "…"
			}

			runesA := []rune(artist)
			if len(runesA) > artistWidth {
				artist = string(runesA[:artistWidth-1]) + "…"
			}

			if !t.hideAlbum {
				runesAl := []rune(album)
				if len(runesAl) > albumWidth {
					album = string(runesAl[:albumWidth-1]) + "…"
				}
			}

			seq := fmt.Sprintf("%*d", seqWidth, idx+1)
			dur := ""
			if tr.Duration > 0 {
				dur = common.FmtDuration(tr.Duration)
			}
			durFmt := fmt.Sprintf("%*s", durWidth, dur)

			prefix := "   "
			if t.active && t.cursor == idx {
				prefix = common.Cursor
			}
			var line string
			if t.hideAlbum {
				line = fmt.Sprintf("%s %s%-*s %-*s %s", seq, prefix, titleWidth, title, artistWidth, artist, durFmt)
			} else {
				line = fmt.Sprintf("%s %s%-*s %-*s %-*s %s", seq, prefix, titleWidth, title, artistWidth, artist, albumWidth, album, durFmt)
			}
			if t.active && t.cursor == idx {
				line = common.CursorStyle.Render(line)
			}
			lines = append(lines, line)
		}
		return strings.Join(lines, "\n")
	}

	// local
	if len(t.entries) == 0 {
		if t.searchMode {
			return strings.Join(lines, "\n")
		}
		style := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center)
		if t.subdirCount > 0 {
			msg := fmt.Sprintf("Contains %d subdirector", t.subdirCount)
			if t.subdirCount == 1 {
				msg += "y"
			} else {
				msg += "ies"
			}
			return style.Render(msg)
		}
		return style.Render("No tracks in this directory")
	}

	lines = append(lines, header, sep)

	for i, e := range t.entries {
		idx := t.offset + i
		title := e.Title
		if title == "" {
			title = e.Name
		}
		artist := e.Artist
		album := e.Album
		runesT := []rune(title)
		if len(runesT) > titleWidth {
			title = string(runesT[:titleWidth-1]) + "…"
		}
		runesA := []rune(artist)
		if len(runesA) > artistWidth {
			artist = string(runesA[:artistWidth-1]) + "…"
		}
		runesAl := []rune(album)
		if len(runesAl) > albumWidth {
			album = string(runesAl[:albumWidth-1]) + "…"
		}
		seq := fmt.Sprintf("%*d", seqWidth, idx+1)
		dur := ""
		if e.Duration > 0 {
			dur = common.FmtDuration(e.Duration)
		}
		durFmt := fmt.Sprintf("%*s", durWidth, dur)
		playing := t.playingPath != "" && e.Path == t.playingPath
		prefix := "   "
		if playing {
			prefix = common.Play + "  "
		} else if t.active && t.cursor == idx {
			prefix = common.Cursor
		}
		line := fmt.Sprintf("%s %s%-*s %-*s %-*s %s", seq, prefix, titleWidth, title, artistWidth, artist, albumWidth, album, durFmt)
		if playing {
			line = common.PlayingTrackStyle.Render(line)
		} else if t.active && t.cursor == idx {
			line = common.CursorStyle.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (t Tracks) View() string {
	content := t.buildBlock()
	pos := t.cursor + 1
	if t.total == 0 {
		pos = 0
	}
	info := []string{fmt.Sprintf("%d of %d", pos, t.total)}
	title := t.title
	if title == "" {
		title = "Content"
	}
	return renderer.Render(content, renderer.Config{
		Width:     t.width,
		Height:    t.height,
		Title:     title,
		InfoItems: info,
		Active:    t.active,
	})
}
