package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	RedisAddr string
	RedisDB   int

	AuthToken string // used for communication with the scale
	Password  string // shared admin password

	FrontendPath string

	PrometheusURL      string
	PrometheusUser     string
	PrometheusPassword string
}

func NewConfig() *Config {
	return &Config{
		RedisAddr: getStringEnvDefault("REDIS_ADDR", "localhost:6379"),
		RedisDB:   getIntEnvDefault("REDIS_DB", 0),

		AuthToken: getStringEnvDefault("AUTH_TOKEN", "test"),
		Password:  getStringEnvDefault("PASSWORD", "test"),

		FrontendPath: getStringEnvDefault("FRONTEND_PATH", "./../frontend/build/"),

		PrometheusURL:      getStringEnvDefault("PROMETHEUS_URL", "http://localhost:9090"),
		PrometheusUser:     getStringEnvDefault("PROMETHEUS_USER", "test"),
		PrometheusPassword: getStringEnvDefault("PROMETHEUS_PASSWORD", "test"),
	}
}

func getStringEnvDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	fmt.Printf("Using default value for %s\n", key)
	return defaultValue
}

func getIntEnvDefault(key string, defaultValue int) int {
	if value, ok := os.LookupEnv(key); ok {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}

	fmt.Printf("Using default value for %s\n", key)
	return defaultValue
}
