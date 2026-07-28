package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Client struct {
	creds *Credentials
}

const baseURL = "https://api.spotify.com"

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
	url := baseURL + "/v1/playlists/" + playlistID + "/tracks?limit=" + strconv.Itoa(limit) + "&offset=" + strconv.Itoa(offset)
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

func (cl *Client) doRequest(method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cl.creds.AccessToken)

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
		if err := RefreshAccessToken(cl.creds); err != nil {
			return nil, fmt.Errorf("token refresh failed: %w", err)
		}
		return cl.doRequestOnce(method, url, body)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (cl *Client) doRequestOnce(method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cl.creds.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
