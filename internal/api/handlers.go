package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/djengua/mqtt-ingestor/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

type APIHandlers struct {
	authService *auth.Service
	pool        *pgxpool.Pool
	logger      *slog.Logger
}

func NewAPIHandlers(authService *auth.Service, pool *pgxpool.Pool, logger *slog.Logger) *APIHandlers {
	return &APIHandlers{
		authService: authService,
		pool:        pool,
		logger:      logger,
	}
}

// Auth Handlers

func (h *APIHandlers) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		h.logger.Error("register error", slog.String("error", err.Error()))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *APIHandlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Error("login error", slog.String("error", err.Error()))
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// Device Handlers

type DeviceResponse struct {
	ID           int64      `json:"id"`
	DeviceID     string     `json:"device_id"`
	NodeType     string     `json:"node_type"`
	GatewayID    *string    `json:"gateway_id"`
	LastSeenAt   *time.Time `json:"last_seen_at"`
	LastSeq      *int64     `json:"last_seq"`
	LastStatus   *int       `json:"last_status"`
	LastBatteryV *float64   `json:"last_battery_v"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type ListDevicesResponse struct {
	Devices []DeviceResponse `json:"devices"`
	Total   int64            `json:"total"`
}

func (h *APIHandlers) HandleListDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	limit := 50
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get total count
	var total int64
	err := h.pool.QueryRow(r.Context(), `SELECT COUNT(*) FROM devices`).Scan(&total)
	if err != nil {
		h.logger.Error("count devices error", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Get devices
	rows, err := h.pool.Query(r.Context(), `
		SELECT id, device_id, node_type, gateway_id, last_seen_at, last_seq, last_status, last_battery_v, created_at, updated_at
		FROM devices
		ORDER BY last_seen_at DESC NULLS LAST
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		h.logger.Error("query devices error", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	devices := []DeviceResponse{}
	for rows.Next() {
		var d DeviceResponse
		if err := rows.Scan(
			&d.ID, &d.DeviceID, &d.NodeType, &d.GatewayID,
			&d.LastSeenAt, &d.LastSeq, &d.LastStatus, &d.LastBatteryV,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			h.logger.Error("scan device error", slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		devices = append(devices, d)
	}

	resp := ListDevicesResponse{
		Devices: devices,
		Total:   total,
	}
	writeJSON(w, http.StatusOK, resp)
}

type TelemetryEventResponse struct {
	ID              int64     `json:"id"`
	DeviceID        string    `json:"device_id"`
	Topic           string    `json:"topic"`
	Seq             int64     `json:"seq"`
	NodeType        string    `json:"node_type"`
	GatewayID       *string   `json:"gateway_id"`
	Status          int       `json:"status"`
	TemperatureC    *float64  `json:"temperature_c"`
	HumidityAirPct  *float64  `json:"humidity_air_pct"`
	SoilMoistureRaw *int      `json:"soil_moisture_raw"`
	SoilMoisturePct *float64  `json:"soil_moisture_pct"`
	BatteryV        *float64  `json:"battery_v"`
	DeviceTS        time.Time `json:"device_ts"`
	IngestedAt      time.Time `json:"ingested_at"`
	PayloadJSON     *any      `json:"payload_json"`
}

type ListTelemetryResponse struct {
	Events []TelemetryEventResponse `json:"events"`
	Total  int64                    `json:"total"`
}

func (h *APIHandlers) HandleGetDeviceTelemetry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	deviceID := strings.TrimPrefix(r.URL.Path, "/api/v1/devices/")
	deviceID = strings.TrimSuffix(deviceID, "/telemetry")

	if deviceID == "" {
		writeError(w, http.StatusBadRequest, "device id is required")
		return
	}

	limit := 100
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get total count
	var total int64
	err := h.pool.QueryRow(r.Context(), `
		SELECT COUNT(*)
		FROM telemetry_events t
		JOIN devices d ON t.device_pk = d.id
		WHERE d.device_id = $1
	`, deviceID).Scan(&total)
	if err != nil {
		h.logger.Error("count telemetry error", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Get events
	rows, err := h.pool.Query(r.Context(), `
		SELECT
			t.id, d.device_id, t.topic, t.seq, t.node_type, t.gateway_id,
			t.status, t.temperature_c, t.humidity_air_pct, t.soil_moisture_raw,
			t.soil_moisture_pct, t.battery_v, t.device_ts, t.ingested_at, t.payload_json
		FROM telemetry_events t
		JOIN devices d ON t.device_pk = d.id
		WHERE d.device_id = $1
		ORDER BY t.device_ts DESC
		LIMIT $2 OFFSET $3
	`, deviceID, limit, offset)
	if err != nil {
		h.logger.Error("query telemetry error", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	events := []TelemetryEventResponse{}
	for rows.Next() {
		var e TelemetryEventResponse
		if err := rows.Scan(
			&e.ID, &e.DeviceID, &e.Topic, &e.Seq, &e.NodeType, &e.GatewayID,
			&e.Status, &e.TemperatureC, &e.HumidityAirPct, &e.SoilMoistureRaw,
			&e.SoilMoisturePct, &e.BatteryV, &e.DeviceTS, &e.IngestedAt, &e.PayloadJSON,
		); err != nil {
			h.logger.Error("scan telemetry error", slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		events = append(events, e)
	}

	resp := ListTelemetryResponse{
		Events: events,
		Total:  total,
	}
	writeJSON(w, http.StatusOK, resp)
}

// Utility functions

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, message string) {
	resp := ErrorResponse{Error: message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}
