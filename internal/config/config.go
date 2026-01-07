package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppEnv    string
	AppPort   string
	DBUser    string
	DBPass    string
	DBName    string
	DBHost    string
	DBPort    string
	JWTSecret string
}

// Load validates that all required environment variables are set.
// It returns an error if any of them are missing.
func Load() (*Config, error) {
	required := []string{
		"APP_ENV",
		"APP_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"DB_HOST",
		"DB_PORT",
		"JWT_SECRET",
	}

	missing := []string{}
	values := make(map[string]string)

	for _, key := range required {
		value, ok := os.LookupEnv(key)
		if !ok || value == "" {
			missing = append(missing, key)
		} else {
			values[key] = value
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missing)
	}

	cfg := &Config{
		AppEnv:    values["APP_ENV"],
		AppPort:   values["APP_PORT"],
		DBUser:    values["DB_USER"],
		DBPass:    values["DB_PASSWORD"],
		DBName:    values["DB_NAME"],
		DBHost:    values["DB_HOST"],
		DBPort:    values["DB_PORT"],
		JWTSecret: values["JWT_SECRET"],
	}

	return cfg, nil
}
