package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/djengua/mqtt-ingestor/internal/api"
	"github.com/djengua/mqtt-ingestor/internal/auth"
	"github.com/djengua/mqtt-ingestor/internal/config"
	"github.com/djengua/mqtt-ingestor/internal/db"
	"github.com/djengua/mqtt-ingestor/internal/httpserver"
	"github.com/djengua/mqtt-ingestor/internal/ingest"
	mqttclient "github.com/djengua/mqtt-ingestor/internal/mqtt"
	"github.com/djengua/mqtt-ingestor/internal/observability"
)

func Run() error {
	logger := observability.NewLogger()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg.PostgresDSN)
	if err != nil {
		return err
	}
	defer pool.Close()

	// Initialize auth services
	userRepo := auth.NewPostgresUserRepository(pool)
	authService := auth.NewService(userRepo, cfg.JWTSecret, 24*time.Hour)

	// Initialize ingest services
	repo := ingest.NewPostgresRepository(pool)
	svc := ingest.NewService(repo, logger)

	// Initialize API handlers
	apiHandlers := api.NewAPIHandlers(authService, pool, logger)

	mqttc := mqttclient.NewClient(
		cfg.MQTTBroker,
		cfg.MQTTClientID,
		cfg.MQTTUsername,
		cfg.MQTTPassword,
		cfg.MQTTTopics,
		cfg.MQTTQoS,
		logger,
		svc.HandleMessage,
	)

	logger.Info("starting mqtt client")
	if err := mqttc.Start(ctx); err != nil {
		return err
	}
	logger.Info("mqtt client started")

	httpSrv := httpserver.New(cfg.HTTPPort, logger, mqttc.IsConnected, apiHandlers, authService)

	logger.Info("starting http server", slog.String("port", cfg.HTTPPort))
	httpSrv.Start()

	logger.Info("service started",
		slog.String("app", cfg.AppName),
		slog.String("env", cfg.AppEnv),
	)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("shutdown requested")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return httpSrv.Shutdown(shutdownCtx)
}
