package events

import (
	"time"
)

type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

type Event struct {
	ID                string    `json:"id"`
	SpacecraftiD      int64     `json:"spacecraft_id"`
	TelemetrySampleID *int64    `json:"telemetry_sample_id,omitempty"`
	Code              string    `json:"code"`
	Severity          Severity  `json:"severity"`
	Message           string    `json:"message"`
	OccuredAt         time.Time `json:"occurred_at"`
	CreatedAt         time.Time `json:"created_at"`
}
