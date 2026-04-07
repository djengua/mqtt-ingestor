package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	repo := ingest.NewPostgresRepository(pool)
	svc := ingest.NewService(repo, logger)

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

	if err := mqttc.Start(ctx); err != nil {
		return err
	}
	defer mqttc.Stop()

	httpSrv := httpserver.New(cfg.HTTPPort, logger, mqttc.IsConnected)
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
