package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const appName = "weazlfeed"

type Config struct {
	ActiveProvider string              `json:"active_provider"`
	Providers      map[string]Provider `json:"providers"`
	Database       Database            `json:"database"`
	UI             UI                  `json:"ui"`
	Feeds          []SeedFeed          `json:"feeds,omitempty"`
}

type Provider struct {
	Type          string `json:"type"`
	ServerURL     string `json:"server_url"`
	Model         string `json:"model"`
	APIKey        string `json:"api_key,omitempty"`
	ContextWindow int    `json:"context_window,omitempty"`
}

type Database struct {
	Path string `json:"path"`
}

type UI struct {
	HideSludge    bool   `json:"hide_sludge"`
	MarkdownStyle string `json:"markdown_style"`
}

type SeedFeed struct {
	Section  string `json:"section,omitempty"`
	Folder   string `json:"folder,omitempty"`
	Category string `json:"category,omitempty"`
	Title    string `json:"title"`
	URL      string `json:"url"`
}

func Load() (Config, string, error) {
	path := configPath()
	cfg, err := LoadPath(path)
	return cfg, path, err
}

func LoadPath(path string) (Config, error) {
	cfg := Default()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return cfg, err
	}
	if err := os.MkdirAll(filepath.Dir(cfg.Database.Path), 0o700); err != nil {
		return cfg, err
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, Save(path, cfg)
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	cfg.withDefaults()
	return cfg, nil
}

func Save(path string, cfg Config) error {
	cfg.withDefaults()
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o600)
}

func Default() Config {
	dataDir := dataDir()
	return Config{
		ActiveProvider: "local-vllm",
		Providers: map[string]Provider{
			"local-vllm": {
				Type:          "vllm",
				ServerURL:     "http://localhost:8000",
				Model:         "local-model",
				ContextWindow: 32768,
			},
			"local-ollama": {
				Type:          "ollama",
				ServerURL:     "http://localhost:11434",
				Model:         "llama3.1",
				ContextWindow: 32768,
			},
		},
		Database: Database{Path: filepath.Join(dataDir, "weazlfeed.sqlite3")},
		UI:       UI{MarkdownStyle: "dark"},
		Feeds:    DefaultFeeds(),
	}
}

func (c *Config) Active() Provider {
	if c.Providers == nil {
		return Provider{}
	}
	return c.Providers[c.ActiveProvider]
}

func (c *Config) withDefaults() {
	def := Default()
	if c.ActiveProvider == "" {
		c.ActiveProvider = def.ActiveProvider
	}
	if len(c.Providers) == 0 {
		c.Providers = def.Providers
	}
	for name, provider := range c.Providers {
		if provider.ContextWindow <= 0 {
			provider.ContextWindow = 32768
			c.Providers[name] = provider
		}
	}
	if c.Database.Path == "" {
		c.Database.Path = def.Database.Path
	}
	if c.UI.MarkdownStyle == "" {
		c.UI.MarkdownStyle = def.UI.MarkdownStyle
	}
	if c.Feeds == nil {
		c.Feeds = def.Feeds
	}
}

func configPath() string {
	if p := os.Getenv("WEAZLFEED_CONFIG"); p != "" {
		return p
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, appName, "config.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", appName, "config.json")
}

func dataDir() string {
	if p := os.Getenv("WEAZLFEED_DATA"); p != "" {
		return p
	}
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, appName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", appName)
}
