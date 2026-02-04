package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"axe-desktop/pkg/models"
	"github.com/joho/godotenv"
)

type Config struct {
	DBPath           string             `json:"db_path"`
	Providers        []models.Provider  `json:"providers"`
	MCPServers       []models.MCPServer `json:"mcp_servers"`
	ActiveProviderID string             `json:"active_provider_id"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".axe-desktop")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	cfg := &Config{
		DBPath: filepath.Join(configDir, "axe-desktop.db"),
		Providers: []models.Provider{
			{
				ID:      "default-gemini",
				Name:    "Google Gemini",
				Type:    models.ProviderGemini,
				Model:   "gemini-2.0-flash",
				Enabled: true,
			},
		},
		MCPServers: []models.MCPServer{
			{
				ID:      "exa",
				Name:    "Exa Search",
				Type:    models.MCPServerHTTP,
				URL:     "https://mcp.exa.ai/mcp",
				Enabled: true,
			},
		},
		ActiveProviderID: "default-gemini",
	}

	configPath := filepath.Join(configDir, "config.json")
	if data, err := os.ReadFile(configPath); err == nil {
		_ = json.Unmarshal(data, cfg)
	}

	// Environment variable overrides for BYOK convenience
	googleKey := os.Getenv("GOOGLE_API_KEY")
	if googleKey == "" {
		googleKey = os.Getenv("AXE_API_KEY")
	}

	for i := range cfg.Providers {
		if cfg.Providers[i].Type == models.ProviderGemini && googleKey != "" {
			cfg.Providers[i].APIKey = googleKey
		}
	}

	return cfg, nil
}

func (c *Config) Save() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".axe-desktop")
	configPath := filepath.Join(configDir, "config.json")

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func (c *Config) GetActiveProvider() *models.Provider {
	for i := range c.Providers {
		if c.Providers[i].ID == c.ActiveProviderID {
			return &c.Providers[i]
		}
	}
	if len(c.Providers) > 0 {
		return &c.Providers[0]
	}
	return nil
}
