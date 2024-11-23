package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
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

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
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
	router.HandleFunc("/api/scale/dashboard", hr.scaleDashboardHandler())
	router.HandleFunc("/api/scale/warehouse", hr.scaleWarehouseHandler())

	router.HandleFunc("/api/pub/active_keg", hr.activeKegHandler())

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
func StartServer(router *mux.Router, port int, logger *logrus.Logger) *http.Server {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Infof("listen: %s", err)
		}
	}()
	logger.Infof("Server Started on port %d", port)

	return srv
}

type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{w, http.StatusOK}
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
