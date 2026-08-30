package telemetry

import (
	"time"
)

type Sample struct {
	ID                int64     `json:"id"`
	SpacecraftID      int64     `json:"spacecraft_id"`
	SequenceNumber    int64     `json:"sequence_number"`
	SourceTimestamp   time.Time `json:"source_timestamp"`
	ReceivedAt        time.Time `json:"received_at"`
	BatteryVoltage    float64   `json:"battery_voltage"`
	BatterySOCPercent float64   `json:"battery_soc_percent"`
	TemperatureC      float64   `json:"temperature_c"`
	Mode              string    `json:"mode"`
}

type CreateSampleParams struct {
	SpacecraftID      int64
	SequenceNumber    int64
	SourceTimestamp   time.Time
	BatteryVoltage    float64
	BatterySOCPercent float64
	TemperatureC      float64
	Mode              string
}
