package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	MusicDir string `json:"music_dir"`
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot get home dir: %w", err)
	}
	return filepath.Join(home, ".config", "resonance", "config.json"), nil
}

func ConfigExists() bool {
	p, err := configPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(p)
	return err == nil
}

func LoadConfig() (Config, error) {
	var cfg Config
	p, err := configPath()
	if err != nil {
		return cfg, err
	}
	f, err := os.Open(p)
	if err != nil {
		return cfg, err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&cfg)
	return cfg, err
}

func SaveConfig(cfg Config) error {
	p, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return fmt.Errorf("cannot create config dir: %w", err)
	}
	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("cannot create config file: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	return enc.Encode(cfg)
}

func ValidateMusicDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", path)
		}
		return fmt.Errorf("cannot access %s: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("directory not readable: %w", err)
	}
	f.Close()
	return nil
}

func GetMusicDir() (string, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return "", err
	}
	if cfg.MusicDir == "" {
		return "", fmt.Errorf("music_dir not configured")
	}
	return cfg.MusicDir, nil
}
