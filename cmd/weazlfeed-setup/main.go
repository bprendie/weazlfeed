package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/bprendie/weazlfeed/internal/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "setup: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	reader := bufio.NewReader(os.Stdin)
	cfg, cfgPath, err := config.Load()
	if err != nil {
		return err
	}
	fmt.Println("WeazlFeed provider setup")
	active := cfg.Active()
	defaultType := firstText(active.Type, "vllm")
	providerType := askChoice(reader, "Provider", []string{"vllm", "ollama"}, defaultType)
	defaultURL := "http://localhost:8000"
	if providerType == "ollama" {
		defaultURL = "http://localhost:11434"
	}
	if active.Type == providerType && active.ServerURL != "" {
		defaultURL = active.ServerURL
	}
	fmt.Println(urlHelp(providerType))
	serverURL := normalizeServerURL(providerType, askString(reader, "Base URL", defaultURL))
	fmt.Printf("Using base URL: %s\n", serverURL)
	models, err := fetchModels(providerType, serverURL)
	var model string
	defaultModelName := defaultModel(providerType)
	if active.Type == providerType && active.Model != "" {
		defaultModelName = active.Model
	}
	if err != nil {
		fmt.Printf("Could not query models: %v\n", err)
		model = askString(reader, "Model name", defaultModelName)
	} else if len(models) == 0 {
		fmt.Println("Provider returned no models.")
		model = askString(reader, "Model name", defaultModelName)
	} else {
		model = askModel(reader, models, defaultModelName)
	}
	cfg = writeProvider(cfg, providerType, serverURL, model, askContextWindow(reader, active.ContextWindow))
	if askChoice(reader, "Add a starter feed", []string{"no", "yes"}, "no") == "yes" {
		title := askString(reader, "Feed title", "")
		url := askString(reader, "Feed URL", "")
		if url != "" {
			cfg.Feeds = append(cfg.Feeds, config.SeedFeed{Title: title, URL: url})
		}
	}
	if err := config.Save(cfgPath, cfg); err != nil {
		return err
	}
	fmt.Printf("Wrote config: %s\n", cfgPath)
	return nil
}

func firstText(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func writeProvider(cfg config.Config, providerType, serverURL, model string, contextWindow int) config.Config {
	if cfg.Providers == nil {
		cfg.Providers = map[string]config.Provider{}
	}
	if contextWindow <= 0 {
		contextWindow = 32768
	}
	id := "primary-" + providerType
	cfg.ActiveProvider = id
	cfg.Providers[id] = config.Provider{
		Type:          providerType,
		ServerURL:     normalizeServerURL(providerType, serverURL),
		Model:         model,
		ContextWindow: contextWindow,
	}
	return cfg
}
