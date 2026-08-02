package spotify

// `json:<key>` these tags maps YOUR field name to THEIR key

type UserProfile struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Product     string `json:"product"`
	Country     string `json:"country"`
	Images      []struct {
		URL string `json:"url"`
	} `json:"images"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// Pagination
type PagingObject[T any] struct {
	Items  []T    `json:"items"`
	Total  int    `json:"total"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Next   string `json:"next"`
}

// Core types
type SpotifyImage struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type SpotifyArtist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SpotifyAlbum struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Images      []SpotifyImage  `json:"images"`
	Artists     []SpotifyArtist `json:"artists"`
	TotalTracks int             `json:"total_tracks"`
}

type SpotifyTrack struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	DurationMs int    `json:"duration_ms"`
	// PlaylistName SpotifyPlaylist `json:"name"`
	Album   SpotifyAlbum    `json:"album"`
	Artists []SpotifyArtist `json:"artists"`
}

type SpotifyPlaylist struct {
	ID          string                          `json:"id"`
	Name        string                          `json:"name"`
	Description string                          `json:"description"`
	Public      *bool                           `json:"public"`
	Images      []SpotifyImage                  `json:"images"`
	Tracks      PagingObject[PlaylistTrackItem] `json:"tracks"`
}

type PlaylistTrackItem struct {
	Track SpotifyTrack `json:"track"`
}

type PlayHistorwy struct {
	Track    SpotifyTrack `json:"track"`
	PlayedAt string       `json:"played_at"`
}

// Saved item wrappers
type SavedTrack struct {
	Track SpotifyTrack `json:"track"`
}
type SavedAlbum struct {
	Album SpotifyAlbum `json:"album"`
}
