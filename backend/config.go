package main

import "os"

type Config struct {
	AuthToken string
}

func NewConfig() *Config {
	return &Config{
		AuthToken: os.Getenv("AUTH_TOKEN"),
	}
}
