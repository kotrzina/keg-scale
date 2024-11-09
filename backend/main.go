package main

import (
	"context"
	"os"

	"github.com/joho/godotenv"
	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/kotrzina/keg-scale/pkg/hook"
	"github.com/kotrzina/keg-scale/pkg/promector"
	"github.com/kotrzina/keg-scale/pkg/prometheus"
	"github.com/kotrzina/keg-scale/pkg/scale"
	"github.com/kotrzina/keg-scale/pkg/store"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := createLogger()

	// for development purposes
	err := godotenv.Load(".env")
	if err != nil {
		// we don't really care if it fails
		logger.Debugf("could not load .env file")
	}
	conf := config.NewConfig()

	c := context.Background()
	ctx, cancel := context.WithCancel(c)

	discord := hook.New(conf.DiscordOpenHook, conf.DiscordKegHook, logger)
	monitor := prometheus.New()
	storage := store.NewRedisStore(conf)
	kegScale := scale.New(ctx, monitor, storage, discord, logger)

	prometheusCollector := promector.NewPromector(
		ctx,
		conf.PrometheusURL,
		conf.PrometheusUser,
		conf.PrometheusPassword,
		kegScale,
		logger,
	)

	StartServer(NewRouter(&HandlerRepository{
		scale:     kegScale,
		promector: prometheusCollector,
		config:    conf,
		monitor:   monitor,
		logger:    logger,
	}), 8080, cancel)
}

func createLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	return logger
}
