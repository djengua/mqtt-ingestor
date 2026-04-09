package ingest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	SaveEvent(ctx context.Context, e Event) error
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) SaveEvent(ctx context.Context, e Event) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var devicePK int64
	err = tx.QueryRow(ctx, `
		INSERT INTO devices (
			device_id,
			node_type,
			gateway_id,
			last_seen_at,
			last_seq,
			last_status,
			last_battery_v,
			updated_at
		)
		VALUES ($1, $2, $3, NOW(), $4, $5, $6, NOW())
		ON CONFLICT (device_id)
		DO UPDATE SET
			node_type = EXCLUDED.node_type,
			gateway_id = EXCLUDED.gateway_id,
			last_seen_at = EXCLUDED.last_seen_at,
			last_seq = EXCLUDED.last_seq,
			last_status = EXCLUDED.last_status,
			last_battery_v = EXCLUDED.last_battery_v,
			updated_at = NOW()
		RETURNING id
	`, e.DeviceID, e.NodeType, e.GatewayID, e.Seq, e.Status, e.BatteryV).Scan(&devicePK)
	if err != nil {
		return fmt.Errorf("upsert device: %w", err)
	}

	var payloadJSON any
	if len(e.PayloadJSON) > 0 {
		if err := json.Unmarshal(e.PayloadJSON, &payloadJSON); err != nil {
			return fmt.Errorf("unmarshal payload json: %w", err)
		}
	}

	_, err = tx.Exec(ctx, `
	INSERT INTO telemetry_events (
		device_pk,
		topic,
		seq,
		node_type,
		gateway_id,
		status,
		temperature_c,
		humidity_air_pct,
		soil_moisture_raw,
		soil_moisture_pct,
		battery_v,
		device_ts,
		boot_id,
		device_ms,
		payload_raw,
		payload_json
	)
	VALUES (
		$1, $2, $3, $4, $5, $6,
		$7, $8, $9, $10, $11, $12,
		$13, $14, $15, $16
	)
	ON CONFLICT (device_pk, boot_id, device_ms) DO NOTHING
`, devicePK, e.Topic, e.Seq, e.NodeType, e.GatewayID, e.Status,
		e.TemperatureC, e.HumidityAirPct, e.SoilMoistureRaw, e.SoilMoisturePct,
		e.BatteryV, e.DeviceTS, e.BootID, e.DeviceMS, e.PayloadRaw, payloadJSON)
	if err != nil {
		return fmt.Errorf("insert telemetry event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

var _ Repository = (*PostgresRepository)(nil)
