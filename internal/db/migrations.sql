CREATE TABLE IF NOT EXISTS devices (
  id BIGSERIAL PRIMARY KEY,
  device_key TEXT NOT NULL UNIQUE,
  last_topic TEXT,
  last_seen_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS telemetry_events (
  id BIGSERIAL PRIMARY KEY,
  device_id BIGINT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
  topic TEXT NOT NULL,
  payload_raw TEXT NOT NULL,
  payload_json JSONB,
  received_at TIMESTAMPTZ NOT NULL,
  ingested_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_telemetry_events_device_received ON telemetry_events(device_id, received_at DESC);

CREATE INDEX IF NOT EXISTS idx_telemetry_events_payload_json ON telemetry_events USING GIN (payload_json);