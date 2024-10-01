package main

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	// for development purposes
	// we don't care about errors here
	_ = godotenv.Load(".env")
	config := NewConfig()

	c := context.Background()
	ctx, cancel := context.WithCancel(c)

	logger := createLogger()
	monitor := NewMonitor()

	store := NewRedisStore(config)

	scale := NewScale(monitor, store, logger, ctx)
	StartServer(NewRouter(&HandlerRepository{
		scale:   scale,
		config:  config,
		monitor: monitor,
		logger:  logger,
	}), 8080, cancel)
}

func createLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	return logger
}
