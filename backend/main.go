package main

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/kotrzina/keg-scale/pkg/promector"
	"github.com/sirupsen/logrus"
	"os"

	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/kotrzina/keg-scale/pkg/prometheus"
	"github.com/kotrzina/keg-scale/pkg/scale"
	"github.com/kotrzina/keg-scale/pkg/store"
)

func main() {
	// for development purposes
	// we don't care about errors here
	_ = godotenv.Load(".env")
	conf := config.NewConfig()

	c := context.Background()
	ctx, cancel := context.WithCancel(c)

	logger := createLogger()
	mon := prometheus.NewMonitor()

	storage := store.NewRedisStore(conf)
	kegScale := scale.NewScale(mon, storage, logger, ctx)

	promector := promector.NewPromector(
		conf.PrometheusUrl,
		conf.PrometheusUser,
		conf.PrometheusPassword,
		kegScale,
		logger,
		ctx,
	)

	StartServer(NewRouter(&HandlerRepository{
		scale:     kegScale,
		promector: promector,
		config:    conf,
		monitor:   mon,
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
