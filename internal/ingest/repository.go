package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

	var deviceID int64
	err = tx.QueryRow(ctx, `
        INSERT INTO devices (device_key, last_topic, last_seen_at, updated_at)
        VALUES ($1, $2, $3, NOW())
        ON CONFLICT (device_key)
        DO UPDATE SET
            last_topic = EXCLUDED.last_topic,
            last_seen_at = EXCLUDED.last_seen_at,
            updated_at = NOW()
        RETURNING id
    `, e.DeviceKey, e.Topic, e.ReceivedAt).Scan(&deviceID)
	if err != nil {
		return fmt.Errorf("upsert device: %w", err)
	}

	var payload any
	var payloadJSON any
	if len(e.PayloadJSON) > 0 {
		if err := json.Unmarshal(e.PayloadJSON, &payload); err == nil {
			payloadJSON = payload
		}
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO telemetry_events (device_id, topic, payload_raw, payload_json, received_at)
        VALUES ($1, $2, $3, $4, $5)
    `, deviceID, e.Topic, e.PayloadRaw, payloadJSON, e.ReceivedAt)
	if err != nil {
		return fmt.Errorf("insert telemetry event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

var _ Repository = (*PostgresRepository)(nil)

func NowUTC() time.Time {
	return time.Now().UTC()
}
