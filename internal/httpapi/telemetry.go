package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/LukeCuzzetto/sentinelops/internal/telemetry"
)

type TelemetryRepository interface {
	CreateSample(
		ctx context.Context,
		params telemetry.CreateSampleParams,
	) (telemetry.Sample, error)

	ListSamplesBySpacecraftID(
		ctx context.Context,
		spacecraftID int64,
	) ([]telemetry.Sample, error)
	GetLatestSampleBySpacecraftID(
		ctx context.Context,
		spacecraftID int64,
	) (telemetry.Sample, error)
}

type createTelemetryRequest struct {
	SequenceNumber    int64     `json:"sequence_number"`
	SourceTimestamp   time.Time `json:"source_timestamp"`
	BatteryVoltage    float64   `json:"battery_voltage"`
	BatterySOCPercent float64   `json:"battery_soc_percent"`
	TemperatureC      float64   `json:"temperature_c"`
	Mode              string    `json:"mode"`
}

func (app *Application) telemetryHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		app.createTelemetryHandler(w, r)

	case http.MethodGet:
		app.listTelemetryHandler(w, r)

	default:
		w.Header().Set(
			"Allow",
			"GET, POST",
		)

		_ = writeJSON(
			w,
			http.StatusMethodNotAllowed,
			errorResponse{
				Error: "method not allowed",
			},
		)
	}
}

func (app *Application) createTelemetryHandler(w http.ResponseWriter, r *http.Request) {
	idValue := r.PathValue("id")

	spacecraftID, err := strconv.ParseInt(idValue, 10, 64)

	if err != nil || spacecraftID < 1 {
		_ = writeJSON(
			w,
			http.StatusBadRequest,
			errorResponse{
				Error: "invalid spacecraft id",
			},
		)
		return

	}

	var input createTelemetryRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		_ = writeJSON(
			w,
			http.StatusBadRequest,
			errorResponse{
				Error: "invalid request body",
			},
		)

		return
	}

	input.Mode = strings.TrimSpace(input.Mode)

	if input.SequenceNumber < 0 {
		_ = writeJSON(
			w,
			http.StatusBadRequest,
			errorResponse{
				Error: "sequence_number must be zero or greater",
			},
		)

		return
	}

	if input.SourceTimestamp.IsZero() {
		_ = writeJSON(
			w,
			http.StatusBadRequest,
			errorResponse{
				Error: "source_timestamp is required",
			},
		)

		return
	}

	if input.BatteryVoltage < 0 {
		_ = writeJSON(
			w,
			http.StatusBadRequest,
			errorResponse{
				Error: "battery_voltage must be zero or greater",
			},
		)

		return
	}

	if input.BatterySOCPercent < 0 || input.BatterySOCPercent > 100 {
		_ = writeJSON(
			w,
			http.StatusBadRequest,
			errorResponse{
				Error: "battery_soc_percent must be between 0 and 100",
			},
		)

		return
	}

	if input.Mode == "" {
		_ = writeJSON(
			w,
			http.StatusBadRequest,
			errorResponse{
				Error: "mode is required",
			},
		)

		return
	}

	params := telemetry.CreateSampleParams{
		SpacecraftID:      spacecraftID,
		SequenceNumber:    input.SequenceNumber,
		SourceTimestamp:   input.SourceTimestamp,
		BatteryVoltage:    input.BatteryVoltage,
		BatterySOCPercent: input.BatterySOCPercent,
		TemperatureC:      input.TemperatureC,
		Mode:              input.Mode,
	}

	created, err := app.telemetryRepository.CreateSample(r.Context(), params)

	if err != nil {
		app.logger.Printf("create telemetry sample failed: %v", err)

		_ = writeJSON(
			w,
			http.StatusInternalServerError,
			errorResponse{
				Error: "internal server error",
			},
		)
		return
	}

	_ = writeJSON(
		w,
		http.StatusCreated,
		created,
	)
}

func (app *Application) listTelemetryHandler(w http.ResponseWriter, r *http.Request) {
	idValue := r.PathValue("id")

	spacecraftID, err := strconv.ParseInt(idValue, 10, 64)

	if err != nil || spacecraftID < 1 {
		_ = writeJSON(
			w,
			http.StatusBadRequest,
			errorResponse{
				Error: "invalid spacecraft id",
			},
		)
		return

	}

	samples, err := app.telemetryRepository.ListSamplesBySpacecraftID(r.Context(), spacecraftID)

	if err != nil {
		app.logger.Printf("list telemetry samples failed: %v", err)

		_ = writeJSON(
			w,
			http.StatusInternalServerError,
			errorResponse{
				Error: "internal server error",
			},
		)
		return
	}

	_ = writeJSON(
		w,
		http.StatusOK,
		samples,
	)
}

func (app *Application) latestTelemetryHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		w.Header().Set(
			"Allow",
			http.MethodGet,
		)

		_ = writeJSON(
			w,
			http.StatusMethodNotAllowed,
			errorResponse{
				Error: "method not allowed",
			},
		)
		return
	}

	idValue := r.PathValue("id")

	spacecraftID, err := strconv.ParseInt(idValue, 10, 64)

	if err != nil || spacecraftID < 1 {
		_ = writeJSON(
			w,
			http.StatusBadRequest,
			errorResponse{
				Error: "invalid spacecraft id",
			},
		)
		return
	}

	latest, err := app.telemetryRepository.GetLatestSampleBySpacecraftID(r.Context(), spacecraftID)

	if errors.Is(err, telemetry.ErrNotFound) {
		_ = writeJSON(
			w,
			http.StatusNotFound,
			errorResponse{
				Error: "no telemetry samples found for spacecraft",
			},
		)
		return
	}

	if err != nil {
		app.logger.Printf("get latest telemetry sample failed: %v", err)

		_ = writeJSON(
			w,
			http.StatusInternalServerError,
			errorResponse{
				Error: "internal server error",
			},
		)
		return
	}

	_ = writeJSON(
		w,
		http.StatusOK,
		latest,
	)
}
