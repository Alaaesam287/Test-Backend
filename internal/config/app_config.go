package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const CONFIG_FILE_PATH = "./internal/config/config.json"


type RateLimitConfig struct {
	RequestsPerSecond      int `json:"requests_per_second"`
	Burst                  int `json:"burst"`
	CleanupIntervalMinutes int `json:"cleanup_interval_minutes"`
}

type AppConfig struct {
	RateLimit RateLimitConfig `json:"rate_limit"`
}

func LoadAppConfig(path string) (*AppConfig, error) {
	if path != CONFIG_FILE_PATH {
		return nil, fmt.Errorf("config path not allowed")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg AppConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}

	if cfg.RateLimit.RequestsPerSecond <= 0 {
		return nil, fmt.Errorf("invalid rate limit config")
	}

	return &cfg, nil
}

func (r RateLimitConfig) CleanupInterval() time.Duration {
	return time.Duration(r.CleanupIntervalMinutes) * time.Minute
}
