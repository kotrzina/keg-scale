package main

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
}

func NewConfig() *Config {
	return &Config{
		RedisAddr: getStringEnvDefault("REDIS_ADDR", "localhost:6379"),
		RedisDB:   getIntEnvDefault("REDIS_DB", 0),

		AuthToken: getStringEnvDefault("AUTH_TOKEN", "test"),
		Password:  getStringEnvDefault("PASSWORD", "test"),

		FrontendPath: getStringEnvDefault("FRONTEND_PATH", "./../frontend/build/"),
	}
}

func getStringEnvDefault(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	fmt.Println(fmt.Sprintf("Using default value for %s", key))
	return defaultValue
}

func getIntEnvDefault(key string, defaultValue int) int {
	if value, ok := os.LookupEnv(key); ok {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}

	fmt.Println(fmt.Sprintf("Using default value for %s", key))
	return defaultValue
}
