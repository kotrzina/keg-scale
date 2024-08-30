package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// NewRouter creates a new HTTP router
func NewRouter(hr *HandlerRepository) *http.ServeMux {
	router := http.NewServeMux()
	router.Handle("/metrics", hr.metricsHandler())
	router.Handle("/", hr.homepageHandler())
	router.HandleFunc("/api/scale/keg", hr.scaleValueHandler())
	router.HandleFunc("/api/scale/ping", hr.scalePingHandler())
	router.HandleFunc("/api/scale/status", hr.scaleStatusHandler())

	return router
}

// StartServer starts HTTP server
// It listens for SIGINT and SIGTERM signals and gracefully stops the server
func StartServer(router *http.ServeMux, port int) {
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
