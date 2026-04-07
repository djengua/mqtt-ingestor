package ingest

import "time"

type Event struct {
	DeviceKey   string
	Topic       string
	PayloadRaw  string
	PayloadJSON []byte
	ReceivedAt  time.Time
}
