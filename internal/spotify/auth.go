package spotify

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"time"
)

const (
	scope       = "user-read-private user-read-email user-library-read playlist-read-private playlist-read-collaborative"
	redirectURI = "http://127.0.0.1:8888/callback"
	baseAuthURL = "https://accounts.spotify.com/authorize"
	tokenURL    = "https://accounts.spotify.com/api/token"
)

func GenerateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func GetAuthURL(clientID, state string) string {
	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("response_type", "code")
	params.Add("redirect_uri", redirectURI)
	params.Add("state", state)
	params.Add("scope", scope)
	return baseAuthURL + "?" + params.Encode()
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}

func Authenticate(clientID, clientSecret string) (*Credentials, error) {
	state, err := GenerateState()
	if err != nil {
		return nil, err
	}

	authURL := GetAuthURL(clientID, state)
	fmt.Println("Opening browser for authentication...")
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Could not open browser automatically.\nPlease visit:\n%s\n\n", authURL)
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "State mismatch", http.StatusBadRequest)
			errCh <- fmt.Errorf("state mismatch")
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "No code received", http.StatusBadRequest)
			errCh <- fmt.Errorf("no code in callback")
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<h1>Authenticated!</h1><p>You can close this window.</p>")
		codeCh <- code
	})

	server := &http.Server{
		Addr:    ":8888",
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		return nil, err
	case <-time.After(5 * time.Minute):
		server.Close()
		return nil, fmt.Errorf("authentication timed out after 5 minutes")
	}

	server.Close()

	creds, err := ExchangeCode(code, clientID, clientSecret)
	if err != nil {
		return nil, err
	}

	creds.ClientID = clientID
	creds.ClientSecret = clientSecret

	if err := SaveCredentials(creds); err != nil {
		return nil, fmt.Errorf("authenticated but failed to save credentials: %w", err)
	}

	fmt.Println("Credentials saved successfully!")
	return creds, nil
}

func ExchangeCode(code, clientID, clientSecret string) (*Credentials, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	}

	return doTokenRequest(data)
}

func RefreshAccessToken(c *Credentials) error {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {c.RefreshToken},
		"client_id":     {c.ClientID},
		"client_secret": {c.ClientSecret},
	}

	newCreds, err := doTokenRequest(data)
	if err != nil {
		return err
	}

	c.AccessToken = newCreds.AccessToken
	c.Expiry = newCreds.Expiry
	if newCreds.RefreshToken != "" {
		c.RefreshToken = newCreds.RefreshToken
	}
	return SaveCredentials(c)
}

func doTokenRequest(data url.Values) (*Credentials, error) {
	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed (%d): %s", resp.StatusCode, string(body))
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &Credentials{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
}
