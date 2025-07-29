package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds all configuration for the trading bot
type Config struct {
	Kraken KrakenConfig `json:"kraken"`
	Risk   RiskConfig   `json:"risk"`
}

// KrakenConfig holds Kraken exchange configuration
type KrakenConfig struct {
	APIKey     string `json:"api_key"`
	PrivateKey string `json:"private_key"`
	BaseURL    string `json:"base_url"`
}

// RiskConfig holds risk management configuration
type RiskConfig struct {
	MaxPortfolioRisk float64 `json:"max_portfolio_risk"`
	MaxDailyLoss     float64 `json:"max_daily_loss"`
	MaxPositions     int     `json:"max_positions"`
}

// Load loads configuration from environment variables and config file
func Load() (*Config, error) {
	cfg := &Config{
		Kraken: KrakenConfig{
			APIKey:     os.Getenv("KRAKEN_API_KEY"),
			PrivateKey: os.Getenv("KRAKEN_PRIVATE_KEY"),
			BaseURL:    "https://api.kraken.com",
		},
		Risk: RiskConfig{
			MaxPortfolioRisk: 0.02, // 2% max risk per trade
			MaxDailyLoss:     0.05, // 5% max daily loss
			MaxPositions:     4,    // Max 4 positions at once
		},
	}

	// Load from config file if it exists
	if _, err := os.Stat("config.json"); err == nil {
		file, err := os.Open("config.json")
		if err != nil {
			return nil, fmt.Errorf("failed to open config file: %w", err)
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		if err := decoder.Decode(cfg); err != nil {
			return nil, fmt.Errorf("failed to decode config file: %w", err)
		}
	}

	// Validate required fields
	if cfg.Kraken.APIKey == "" || cfg.Kraken.PrivateKey == "" {
		return nil, fmt.Errorf("KRAKEN_API_KEY and KRAKEN_PRIVATE_KEY environment variables are required")
	}

	return cfg, nil
}
