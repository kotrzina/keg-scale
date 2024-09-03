package main

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AuthToken  string
	BufferSize int
}

func NewConfig() *Config {
	return &Config{
		AuthToken:  getStringEnvDefault("AUTH_TOKEN", "test"),
		BufferSize: getIntEnvDefault("BUFFER_SIZE", 1000),
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
