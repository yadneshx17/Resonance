package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func (cl *Client) LikedSongs() {

}

func (cl *Client) Playlists() {

}

func (cl *Client) PlaylistbyID(playlistID string) {

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
