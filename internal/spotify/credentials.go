package spotify

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/x/term"
)

type Credentials struct {
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
}

func credentialsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot get home dir: %w", err)
	}
	return filepath.Join(home, ".config", "resonance", "credentials", "spotify.json"), nil
}

// Checks if file exists
func CredentialsExist() bool {
	path, err := credentialsPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

func SaveCredentials(c *Credentials) error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("cannot create credentials dir: %w", err)
	}

	// owner only
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("cannot create credentials file: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	return enc.Encode(c)
}

func LoadCredentials() (*Credentials, error) {
	path, err := credentialsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read credentials: %w", err)
	}
	var c Credentials
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("cannot parse credentials: %w", err)
	}
	return &c, nil
}

func readSecret(prompt string) (string, error) {
	fmt.Print(prompt)
	fd := int(syscall.Stdin)
	password, err := term.ReadPassword(uintptr(fd))
	fmt.Println() // newline after hidden input
	if err != nil {
		return "", fmt.Errorf("failed to read secret: %w", err)
	}
	return strings.TrimSpace(string(password)), nil
}

func PromptCredentials() (*Credentials, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Spotify Client ID: ")
	clientID, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read client ID: %w", err)
	}
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}

	// fmt.Print("Enter Spotify Client Secret: ")
	clientSecret, err := readSecret("Enter Spotify Client Secret (Hidden input): ") // hidden
	if err != nil {
		return nil, fmt.Errorf("failed to read client secret: %w", err)
	}
	clientSecret = strings.TrimSpace(clientSecret)
	if clientSecret == "" {
		return nil, fmt.Errorf("client secret cannot be empty")
	}

	return &Credentials{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}

func RemoveCredentials() error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot remove credentials: %w", err)
	}
	return nil
}
