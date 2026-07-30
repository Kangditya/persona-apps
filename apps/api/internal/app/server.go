package app

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const readinessTimeout = 2 * time.Second

type readinessChecker interface {
	PingContext(context.Context) error
}

func NewServer(address string, database readinessChecker, logger *slog.Logger) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeStatus(w, http.StatusOK, "ok")
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
		defer cancel()

		if err := database.PingContext(ctx); err != nil {
			logger.Warn("readiness check failed", "error", err)
			writeStatus(w, http.StatusServiceUnavailable, "unavailable")
			return
		}

		writeStatus(w, http.StatusOK, "ready")
	})

	return &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}
}

func writeStatus(w http.ResponseWriter, code int, status string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = io.WriteString(w, "{\"status\":\""+status+"\"}\n")
}
