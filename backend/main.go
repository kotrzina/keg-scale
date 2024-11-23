package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
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
	c := context.Background()
	ctx, cancel := context.WithCancel(c)

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
	botka := hook.NewBotka(whatsapp, conf, logger)
	discord := hook.New(ctx, conf.DiscordOpenHook, conf.DiscordKegHook, logger)
	monitor := prometheus.New()
	storage := store.NewRedisStore(ctx, conf)
	kegScale := scale.New(ctx, monitor, storage, discord, botka, logger)

	prometheusCollector := promector.NewPromector(
		ctx,
		conf.PrometheusURL,
		conf.PrometheusUser,
		conf.PrometheusPassword,
		kegScale,
		logger,
	)

	router := NewRouter(&HandlerRepository{
		scale:     kegScale,
		promector: prometheusCollector,
		config:    conf,
		monitor:   monitor,
		logger:    logger,
	})

	srv := StartServer(router, 8080, logger)

	<-done
	logger.Infof("Terminate signal received")
	cancel() // cancel application context

	shutdownContext, shutdownContextCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownContextCancel()
	if err := srv.Shutdown(shutdownContext); err != nil {
		logger.Errorf("Server Shutdown Failed:%+v", err)
	}

	logger.Infof("Server Exited")
}

func createLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	return logger
}
