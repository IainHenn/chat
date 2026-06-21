package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	Username string `json:"username"`
	Server   string `json:"server"`
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "chat", "config.json"), nil
}

func Load() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, errors.New("not signed up — run: chat signup <username>")
		}
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.Username == "" {
		return Config{}, errors.New("not signed up — run: chat signup <username>")
	}
	if cfg.Server == "" {
		cfg.Server = "http://localhost:8080"
	}
	return cfg, nil
}

func Save(username, server string) error {
	if server == "" {
		server = "http://localhost:8080"
	}

	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.Marshal(Config{Username: username, Server: server})
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func ServerURL(flag string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	cfg, err := Load()
	if err != nil {
		return "", err
	}
	return cfg.Server, nil
}
