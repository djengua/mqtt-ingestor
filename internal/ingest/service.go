package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type Service struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

func (s *Service) HandleMessage(ctx context.Context, topic string, payload []byte) error {
	deviceKey := extractDeviceKey(topic)
	if deviceKey == "" {
		return fmt.Errorf("unable to extract device key from topic %s", topic)
	}

	raw := string(payload)
	var payloadJSON []byte
	if json.Valid(payload) {
		payloadJSON = payload
	}

	evt := Event{
		DeviceKey:   deviceKey,
		Topic:       topic,
		PayloadRaw:  raw,
		PayloadJSON: payloadJSON,
		ReceivedAt:  time.Now().UTC(),
	}

	if err := s.repo.SaveEvent(ctx, evt); err != nil {
		return err
	}

	s.logger.Info("telemetry ingested",
		slog.String("device_key", evt.DeviceKey),
		slog.String("topic", evt.Topic),
		slog.Time("received_at", evt.ReceivedAt),
	)

	return nil
}

func extractDeviceKey(topic string) string {
	// Ejemplo esperado: sensores/<deviceKey>/telemetry
	parts := strings.Split(topic, "/")
	if len(parts) >= 3 {
		return parts[1]
	}
	return ""
}
