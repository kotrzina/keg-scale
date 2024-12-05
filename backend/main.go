package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/kotrzina/keg-scale/pkg/ai"
	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/kotrzina/keg-scale/pkg/hook"
	"github.com/kotrzina/keg-scale/pkg/promector"
	"github.com/kotrzina/keg-scale/pkg/prometheus"
	"github.com/kotrzina/keg-scale/pkg/scale"
	"github.com/kotrzina/keg-scale/pkg/store"
	"github.com/kotrzina/keg-scale/pkg/wa"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	logger := createLogger()

	// for development purposes
	err := godotenv.Load(".env")
	if err != nil {
		// we don't really care if it fails
		logger.Debugf("could not load .env file")
	}
	conf := config.NewConfig()

	whatsapp := wa.New(ctx, conf, logger)
	defer whatsapp.Close()

	monitor := prometheus.New()
	storage := store.NewRedisStore(ctx, conf)
	kegScale := scale.New(ctx, monitor, storage, conf, whatsapp, logger)
	intelligence := ai.NewAi(ctx, conf, kegScale, monitor, logger)
	_ = hook.NewBotka(whatsapp, kegScale, intelligence, conf, storage, logger)
	prometheusCollector := promector.NewPromector(ctx, conf, kegScale, logger)

	router := NewRouter(&HandlerRepository{
		scale:     kegScale,
		promector: prometheusCollector,
		ai:        intelligence,
		config:    conf,
		monitor:   monitor,
		logger:    logger,
	})

	srv := StartServer(router, 8080, logger)

	<-done
	logger.Infof("Terminate signal received")

	shutdownContext, shutdownContextCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownContextCancel()
	if err := srv.Shutdown(shutdownContext); err != nil {
		logger.Errorf("Server Shutdown Failed:%+v", err)
	}

	cancel() // cancel application context

	logger.Infof("Server Exited")
}

func createLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	return logger
}
