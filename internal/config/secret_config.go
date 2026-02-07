package config

import (
	"fmt"
	"os"
)

type Secret struct {
	AppEnv        string
	AppPort       string
	DBUser        string
	DBPass        string
	DBName        string
	DBHost        string
	DBPort        string
	JWTSecret     string
	MinIOEndpoint string
	MinIOUser     string
	MinIOPass     string
	MinIOBucket   string
}

func LoadSecrets() (*Secret, error) {
	required := []string{
		"APP_ENV",
		"APP_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"DB_HOST",
		"JWT_SECRET",
		"MINIO_ENDPOINT",
		"MINIO_USER",
		"MINIO_PASS",
		"MINIO_BUCKET",
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
		return nil, fmt.Errorf("missing env vars: %v", missing)
	}

	return &Secret{
		AppEnv:        values["APP_ENV"],
		AppPort:       values["APP_PORT"],
		DBUser:        values["DB_USER"],
		DBPass:        values["DB_PASSWORD"],
		DBName:        values["DB_NAME"],
		DBHost:        values["DB_HOST"],
		JWTSecret:     values["JWT_SECRET"],
		MinIOEndpoint: values["MINIO_ENDPOINT"],
		MinIOUser:     values["MINIO_USER"],
		MinIOPass:     values["MINIO_PASS"],
		MinIOBucket:   values["MINIO_BUCKET"],
	}, nil
}
