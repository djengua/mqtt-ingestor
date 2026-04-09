package ingest

import "time"

type Event struct {
	DeviceID        string
	NodeType        string
	GatewayID       string
	BootID          *int64
	DeviceMS        *int64
	Topic           string
	Seq             int64
	Status          int
	TemperatureC    *float64
	HumidityAirPct  *float64
	SoilMoistureRaw *int
	SoilMoisturePct *float64
	BatteryV        *float64
	DeviceTS        time.Time
	PayloadRaw      string
	PayloadJSON     []byte
}
