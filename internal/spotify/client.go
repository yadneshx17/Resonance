package spotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type Client struct {
	creds *Credentials
}

const baseURL = "https://api.spotify.com"

// refreshToken is a package var so tests can stub out the network call.
var refreshToken = RefreshAccessToken

func NewClient(c *Credentials) *Client {
	return &Client{creds: c}
}

func (cl *Client) GetUserProfile() (*UserProfile, error) {
	body, err := cl.doRequest(http.MethodGet, baseURL+"/v1/me", nil)
	if err != nil {
		return nil, err
	}
	var profile UserProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse user profile: %w", err)
	}
	return &profile, nil
}

// / Liked
func (cl *Client) GetSavedTracks(limit, offset int) (*PagingObject[SavedTrack], error) {
	url := baseURL + "/v1/me/tracks?limit=" + strconv.Itoa(limit) + "&offset=" + strconv.Itoa(offset)
	body, err := cl.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result PagingObject[SavedTrack]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse saved tracks: %w", err)
	}
	return &result, nil
}

func (cl *Client) GetSavedAlbums(limit, offset int) (*PagingObject[SavedAlbum], error) {
	url := baseURL + "/v1/me/albums?limit=" + strconv.Itoa(limit) + "&offset=" + strconv.Itoa(offset)
	body, err := cl.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result PagingObject[SavedAlbum]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse saved albums: %w", err)
	}
	return &result, nil
}

func (cl *Client) GetPlaylists(limit, offset int) (*PagingObject[SpotifyPlaylist], error) {
	url := baseURL + "/v1/me/playlists?limit=" + strconv.Itoa(limit) + "&offset=" + strconv.Itoa(offset)
	body, err := cl.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result PagingObject[SpotifyPlaylist]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse playlists: %w", err)
	}
	return &result, nil
}

func (cl *Client) GetPlaylistTracks(playlistID string, limit, offset int) (*PagingObject[PlaylistTrackItem], error) {
	// /tracks is deprecated and can return 403; /items is the current endpoint.
	url := baseURL + "/v1/playlists/" + playlistID + "/items?limit=" + strconv.Itoa(limit) + "&offset=" + strconv.Itoa(offset)
	body, err := cl.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result PagingObject[PlaylistTrackItem]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse playlist tracks: %w", err)
	}
	return &result, nil
}

func (cl *Client) GetAlbumTracks(albumID string, limit, offset int) (*PagingObject[SpotifyTrack], error) {
	url := baseURL + "/v1/albums/" + albumID + "/tracks?limit=" + strconv.Itoa(limit) + "&offset=" + strconv.Itoa(offset)
	body, err := cl.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result PagingObject[SpotifyTrack]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse album tracks: %w", err)
	}
	return &result, nil
}

// --- Playback (Spotify Connect) ---

func (cl *Client) GetPlaybackState() (*PlayerState, error) {
	body, err := cl.doRequest(http.MethodGet, baseURL+"/v1/me/player", nil)
	if err != nil {
		return nil, err
	}
	// 204 No Content means nothing is playing / no active device.
	if len(bytes.TrimSpace(body)) == 0 {
		return nil, nil
	}
	var state PlayerState
	if err := json.Unmarshal(body, &state); err != nil {
		return nil, fmt.Errorf("failed to parse player state: %w", err)
	}
	return &state, nil
}

type devicesResponse struct {
	Devices []SpotifyDevice `json:"devices"`
}

func (cl *Client) GetAvailableDevices() ([]SpotifyDevice, error) {
	body, err := cl.doRequest(http.MethodGet, baseURL+"/v1/me/player/devices", nil)
	if err != nil {
		return nil, err
	}
	var resp devicesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse devices: %w", err)
	}
	return resp.Devices, nil
}

// PlayURI starts playback of a single track on the given device.
func (cl *Client) PlayURI(deviceID, uri string, positionMs int) error {
	payload := map[string]any{
		"uris":        []string{uri},
		"position_ms": positionMs,
	}
	return cl.doJSON(http.MethodPut, baseURL+"/v1/me/player/play?device_id="+url.QueryEscape(deviceID), payload)
}

func (cl *Client) ResumePlayback(deviceID string) error {
	return cl.doJSON(http.MethodPut, baseURL+"/v1/me/player/play?device_id="+url.QueryEscape(deviceID), nil)
}

func (cl *Client) PausePlayback(deviceID string) error {
	return cl.doJSON(http.MethodPut, baseURL+"/v1/me/player/pause?device_id="+url.QueryEscape(deviceID), nil)
}

func (cl *Client) SeekPlayback(deviceID string, positionMs int) error {
	u := baseURL + "/v1/me/player/seek?device_id=" + url.QueryEscape(deviceID) + "&position_ms=" + strconv.Itoa(positionMs)
	return cl.doJSON(http.MethodPut, u, nil)
}

func (cl *Client) SetDeviceVolume(deviceID string, pct int) error {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	u := baseURL + "/v1/me/player/volume?device_id=" + url.QueryEscape(deviceID) + "&volume_percent=" + strconv.Itoa(pct)
	return cl.doJSON(http.MethodPut, u, nil)
}

func (cl *Client) doJSON(method, url string, payload any) error {
	var body []byte
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = b
	}
	_, err := cl.doRequest(method, url, body)
	return err
}

// doRequest performs an API request. The body is passed as bytes so it can be
// replayed unchanged if a 401 forces an access-token refresh before retrying.
func (cl *Client) doRequest(method, url string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cl.creds.AccessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// auto-refresh access token
	if resp.StatusCode == http.StatusUnauthorized {
		if err := refreshToken(cl.creds); err != nil {
			return nil, fmt.Errorf("token refresh failed: %w", err)
		}
		return cl.doRequestOnce(method, url, body)
	}

	// Playback endpoints reply 204 No Content on success.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("request failed (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (cl *Client) doRequestOnce(method, url string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cl.creds.AccessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Playback endpoints reply 204 No Content on success.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("request failed (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
