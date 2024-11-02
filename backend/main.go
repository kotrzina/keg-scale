package main

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"os"

	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/kotrzina/keg-scale/pkg/promector"
	"github.com/kotrzina/keg-scale/pkg/prometheus"
	"github.com/kotrzina/keg-scale/pkg/scale"
	"github.com/kotrzina/keg-scale/pkg/store"
)

func main() {
	// for development purposes
	// we don't care about errors here
	_ = godotenv.Load(".env")
	conf := config.NewConfig()

	logger := createLogger()

	c := context.Background()
	ctx, cancel := context.WithCancel(c)

	monitor := prometheus.NewMonitor()

	storage := store.NewRedisStore(conf)
	kegScale := scale.NewScale(monitor, storage, logger, ctx)

	prometheusCollector := promector.NewPromector(
		conf.PrometheusUrl,
		conf.PrometheusUser,
		conf.PrometheusPassword,
		kegScale,
		logger,
		ctx,
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
