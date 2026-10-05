package events

import (
	"time"
)

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

type Event struct {
	ID                int64     `json:"id"`
	SpacecraftID      int64     `json:"spacecraft_id"`
	TelemetrySampleID *int64    `json:"telemetry_sample_id,omitempty"`
	Code              string    `json:"code"`
	Severity          Severity  `json:"severity"`
	Message           string    `json:"message"`
	OccurredAt        time.Time `json:"occurred_at"`
	CreatedAt         time.Time `json:"created_at"`
}

type CreateEventParams struct {
	SpacecraftID      int64
	TelemetrySampleID *int64
	Code              string
	Severity          Severity
	Message           string
	OccurredAt        time.Time
}
