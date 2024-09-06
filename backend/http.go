package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
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

			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

			if r.Method == "OPTIONS" {
				w.WriteHeader(204)
				return
			}

			lrw := NewLoggingResponseWriter(w)
			handler.ServeHTTP(lrw, r)
			d := time.Since(start)

			hr.logger.WithFields(logrus.Fields{
				"status":     lrw.statusCode,
				"method":     r.Method,
				"path":       r.URL.Path,
				"remoteAddr": r.RemoteAddr,
				"durationMs": d.Milliseconds(),
				"duration":   d.String(),
			}).Info("Request")
		})
	})

	router.Handle("/metrics", hr.metricsHandler())
	router.HandleFunc("/api/scale/push", hr.scaleMessageHandler())
	router.HandleFunc("/api/scale/status", hr.scaleStatusHandler())
	router.HandleFunc("/api/scale/print", hr.scalePrintHandler())
	router.HandleFunc("/api/scale/dashboard", hr.scaleDashboardHandler())

	// frontend
	dir := hr.config.FrontendPath
	router.PathPrefix("/").Handler(http.StripPrefix("/", reactRedirect(http.FileServer(http.Dir(dir)), dir)))

	return router
}

// reactRedirect is a middleware that redirects all requests to the React app (index.html)
// it checks if the requested file exists and if not it redirects to index.html
func reactRedirect(server http.Handler, dir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the requested file exists then return if; otherwise return index.html (file server default page)
		if r.URL.Path != "/" {
			fullPath := dir + strings.TrimPrefix(path.Clean(r.URL.Path), "/")
			_, err := os.Stat(fullPath)
			if err != nil {
				if !os.IsNotExist(err) {
					panic(err)
				}
				// Requested file does not exist so we return the default (resolves to index.html)
				r.URL.Path = "/"
			}
		}

		server.ServeHTTP(w, r)
	})
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

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
