package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

type Service struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

type incomingPayload struct {
	Gateway         string   `json:"gateway"`
	DeviceID        string   `json:"deviceId"`
	NodeType        string   `json:"nodeType"`
	BootID          *int64   `json:"bootId"`
	DeviceMS        *int64   `json:"deviceMs"`
	Seq             int64    `json:"seq"`
	Status          int      `json:"status"`
	TemperatureC    *float64 `json:"temperatureC"`
	HumidityAirPct  *float64 `json:"humidityAirPct"`
	SoilMoistureRaw *int     `json:"soilMoistureRaw"`
	SoilMoisturePct *float64 `json:"soilMoisturePct"`
	BatteryMv       *float64 `json:"batteryMv"`
}

func (s *Service) HandleMessage(ctx context.Context, topic string, payload []byte) error {
	if !json.Valid(payload) {
		return fmt.Errorf("payload is not valid json")
	}

	var in incomingPayload
	if err := json.Unmarshal(payload, &in); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	if in.DeviceID == "" {
		return fmt.Errorf("payload missing deviceId")
	}
	if in.NodeType == "" {
		return fmt.Errorf("payload missing nodeType")
	}
	if in.Seq == 0 {
		s.logger.Warn("payload seq is 0", slog.String("topic", topic))
	}

	var batteryV *float64
	if in.BatteryMv != nil {
		v := *in.BatteryMv / 1000.0
		batteryV = &v
	}

	evt := Event{
		DeviceID:        in.DeviceID,
		NodeType:        in.NodeType,
		GatewayID:       in.Gateway,
		BootID:          in.BootID,
		DeviceMS:        in.DeviceMS,
		Topic:           topic,
		Seq:             in.Seq,
		Status:          in.Status,
		TemperatureC:    in.TemperatureC,
		HumidityAirPct:  in.HumidityAirPct,
		SoilMoistureRaw: in.SoilMoistureRaw,
		SoilMoisturePct: in.SoilMoisturePct,
		BatteryV:        batteryV,
		DeviceTS:        time.Now().UTC(),
		PayloadRaw:      string(payload),
		PayloadJSON:     payload,
	}

	if err := s.repo.SaveEvent(ctx, evt); err != nil {
		return err
	}

	s.logger.Info("telemetry ingested",
		slog.String("device_id", evt.DeviceID),
		slog.String("topic", evt.Topic),
		slog.Int64("seq", evt.Seq),
		slog.Time("device_ts", evt.DeviceTS),
	)

	return nil
}
