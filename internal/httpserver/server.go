package httpserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/djengua/mqtt-ingestor/internal/api"
	"github.com/djengua/mqtt-ingestor/internal/auth"
)

type ReadinessFunc func() bool

type Server struct {
	srv    *http.Server
	logger *slog.Logger
}

func New(port string, logger *slog.Logger, readiness ReadinessFunc, apiHandlers *api.APIHandlers, authService *auth.Service) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if !readiness() {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "not_ready"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": "ready"})
	})

	// Auth endpoints (public)
	mux.HandleFunc("/api/v1/auth/register", apiHandlers.HandleRegister)
	mux.HandleFunc("/api/v1/auth/login", apiHandlers.HandleLogin)

	// Protected endpoints with auth middleware
	authMw := api.AuthMiddleware(authService, logger)

	mux.HandleFunc("/api/v1/devices", func(w http.ResponseWriter, r *http.Request) {
		authMw(http.HandlerFunc(apiHandlers.HandleListDevices)).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/devices/", func(w http.ResponseWriter, r *http.Request) {
		authMw(http.HandlerFunc(apiHandlers.HandleGetDeviceTelemetry)).ServeHTTP(w, r)
	})

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &Server{srv: srv, logger: logger}
}

func (s *Server) Start() {
	go func() {
		s.logger.Info("http server listening", slog.String("addr", s.srv.Addr))
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("http server error", slog.String("error", err.Error()))
		}
	}()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
