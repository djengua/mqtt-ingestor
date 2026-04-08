CREATE TABLE IF NOT EXISTS devices (
  id BIGSERIAL PRIMARY KEY,
  device_id TEXT NOT NULL UNIQUE,
  node_type TEXT NOT NULL,
  gateway_id TEXT,
  last_seen_at TIMESTAMPTZ,
  last_seq BIGINT,
  last_status INTEGER,
  last_battery_v NUMERIC(6, 3),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS telemetry_events (
  id BIGSERIAL PRIMARY KEY,
  device_pk BIGINT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
  topic TEXT NOT NULL,
  seq BIGINT NOT NULL,
  node_type TEXT NOT NULL,
  gateway_id TEXT,
  status INTEGER NOT NULL,
  temperature_c NUMERIC(8, 3),
  humidity_air_pct NUMERIC(8, 3),
  soil_moisture_raw INTEGER,
  soil_moisture_pct NUMERIC(8, 3),
  battery_v NUMERIC(6, 3),
  device_ts TIMESTAMPTZ NOT NULL,
  ingested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  payload_raw TEXT NOT NULL,
  payload_json JSONB
);

CREATE INDEX IF NOT EXISTS idx_devices_device_id ON devices(device_id);

CREATE INDEX IF NOT EXISTS idx_telemetry_events_device_pk_device_ts ON telemetry_events(device_pk, device_ts DESC);

CREATE INDEX IF NOT EXISTS idx_telemetry_events_gateway_id ON telemetry_events(gateway_id);

CREATE INDEX IF NOT EXISTS idx_telemetry_events_node_type ON telemetry_events(node_type);

CREATE INDEX IF NOT EXISTS idx_telemetry_events_status ON telemetry_events(status);

CREATE INDEX IF NOT EXISTS idx_telemetry_events_payload_json ON telemetry_events USING GIN (payload_json);

CREATE UNIQUE INDEX IF NOT EXISTS uq_telemetry_device_seq ON telemetry_events(device_pk, seq);