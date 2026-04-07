package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName  string
	AppEnv   string
	HTTPPort string

	MQTTBroker   string
	MQTTClientID string
	MQTTUsername string
	MQTTPassword string
	MQTTTopics   []string
	MQTTQoS      byte

	PostgresDSN string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	qos, err := parseQoS(getEnv("MQTT_QOS", "1"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid MQTT_QOS: %w", err)
	}

	cfg := Config{
		AppName:      getEnv("APP_NAME", "mqtt-ingestor"),
		AppEnv:       getEnv("APP_ENV", "dev"),
		HTTPPort:     getEnv("HTTP_PORT", "8080"),
		MQTTBroker:   getEnv("MQTT_BROKER", "tcp://localhost:1883"),
		MQTTClientID: getEnv("MQTT_CLIENT_ID", "mqtt-ingestor-1"),
		MQTTUsername: getEnv("MQTT_USERNAME", ""),
		MQTTPassword: getEnv("MQTT_PASSWORD", ""),
		MQTTTopics:   splitCSV(getEnv("MQTT_TOPICS", "sensores/+/telemetry")),
		MQTTQoS:      qos,
		PostgresDSN:  getEnv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/iot?sslmode=disable"),
	}

	if cfg.MQTTBroker == "" || len(cfg.MQTTTopics) == 0 || cfg.PostgresDSN == "" {
		return Config{}, fmt.Errorf("missing required configuration")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(v string) []string {
	raw := strings.Split(v, ",")
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func parseQoS(v string) (byte, error) {
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}
	if n < 0 || n > 2 {
		return 0, fmt.Errorf("qos must be 0, 1 or 2")
	}
	return byte(n), nil
}
