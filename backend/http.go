package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

// NewRouter creates a new HTTP router
func NewRouter(hr *HandlerRepository) *mux.Router {

	router := mux.NewRouter()
	router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			handler.ServeHTTP(w, r)
			d := time.Since(start)

			hr.logger.WithFields(logrus.Fields{
				"method":     r.Method,
				"path":       r.URL.Path,
				"remoteAddr": r.RemoteAddr,
				"durationMs": d.Milliseconds(),
				"duration":   d.String(),
			}).Info("Request")
		})
	})

	router.Handle("/metrics", hr.metricsHandler())
	router.Handle("/", hr.homepageHandler())
	router.HandleFunc("/api/scale/keg", hr.scaleValueHandler())
	router.HandleFunc("/api/scale/ping", hr.scalePingHandler())
	router.HandleFunc("/api/scale/status", hr.scaleStatusHandler())
	router.HandleFunc("/api/scale/print", hr.scalePrintHandler())

	return router
}

// StartServer starts HTTP server
// It listens for SIGINT and SIGTERM signals and gracefully stops the server
func StartServer(router *mux.Router, port int) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("listen: %s\n", err)
		}
	}()
	log.Printf("Server Started on port %d", port)

	<-done
	log.Printf("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	log.Printf("Server Exited Properly")
}
