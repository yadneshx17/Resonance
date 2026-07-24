package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/yadneshx17/resonance/internal/spotify"
	tui "github.com/yadneshx17/resonance/internal/ui"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		tui.Run()
		return
	}

	if args[0] == "spotify" {
		handleSpotify(args[1:])
		return
	}

	fmt.Printf("Unknown command: %s\n\n", args[0])
	printUsage()
	os.Exit(1)
}

func handleSpotify(args []string) {
	if len(args) == 0 {
		printSpotifyHelp()
		return
	}

	switch args[0] {
	case "login":
		spotifyLogin()
	case "logout":
		spotifyLogout()
	case "whoami":
		spotifyWhoami()
	default:
		fmt.Printf("Unknown spotify command: %s\n\n", args[0])
		printSpotifyHelp()
		os.Exit(1)
	}
}

func spotifyLogin() {
	if spotify.CredentialsExist() {
		fmt.Print("Spotify credentials already exist. Re-authenticate? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return
		}
	}

	creds, err := spotify.PromptCredentials()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	authenticatedCreds, err := spotify.Authenticate(creds.ClientID, creds.ClientSecret)
	if err != nil {
		fmt.Printf("Authentication failed: %v\n", err)
		os.Exit(1)
	}

	_ = authenticatedCreds
	fmt.Println("Authentication complete!")
}

func spotifyLogout() {
	if !spotify.CredentialsExist() {
		fmt.Println("Not authenticated. Nothing to logout")
		return

	}

	fmt.Print("Are you sure you want to logout? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		fmt.Println("Aborted.")
		return
	}

	if err := spotify.RemoveCredentials(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Logged out. Credentials removed.")
}

func spotifyWhoami() {
	creds, err := spotify.LoadCredentials()
	if err != nil {
		fmt.Println("Not authenticated. Run: resonance spotify login")
		os.Exit(1)
	}

	client := spotify.NewClient(creds)
	profile, err := client.GetUserProfile()
	if err != nil {
		fmt.Printf("Error fetching profile: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("ID:           %s\n", profile.ID)
	fmt.Printf("Name:         %s\n", profile.DisplayName)
	fmt.Printf("Email:        %s\n", profile.Email)
	fmt.Printf("Subscription: %s\n", profile.Product)
	fmt.Printf("Country:      %s\n", profile.Country)
}

func printUsage() {
	fmt.Println("Usage: resonance [command]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  (none)            Launch the TUI player")
	fmt.Println("  spotify login     Authenticate with Spotify")
	fmt.Println("  spotify logout    De-Authenticate with Spotify")
	fmt.Println("  spotify whoami    Display your Spotify profile")
}

func printSpotifyHelp() {
	fmt.Println("Usage: resonance spotify <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  login    Authenticate with Spotify")
	fmt.Println("  logout   Remove Stored Credentials")
	fmt.Println("  whoami   Display your Spotify profile")
}
