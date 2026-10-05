package events

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	database *pgxpool.Pool
}

func NewRepository(database *pgxpool.Pool) *Repository {
	return &Repository{
		database: database,
	}
}

func (r *Repository) CreateEvent(ctx context.Context, params CreateEventParams) (Event, error) {
	const query = `
	INSERT INTO operational_events (spacecraft_id, telemetry_sample_id, code, severity, message, occurred_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, spacecraft_id, telemetry_sample_id, code, severity, message, occurred_at, created_at
	`

	var event Event

	err := r.database.QueryRow(ctx, query,
		params.SpacecraftID,
		params.TelemetrySampleID,
		params.Code,
		params.Severity,
		params.Message,
		params.OccurredAt,
	).Scan(
		&event.ID,
		&event.SpacecraftID,
		&event.TelemetrySampleID,
		&event.Code,
		&event.Severity,
		&event.Message,
		&event.OccurredAt,
		&event.CreatedAt,
	)

	if err != nil {
		return Event{}, fmt.Errorf("failed to create event: %w", err)
	}

	return event, nil
}

func (r *Repository) ListEventsBySpacecraftID(ctx context.Context, spacecraftID int64) ([]Event, error) {

	const query = `
	SELECT id, spacecraft_id, telemetry_sample_id, code, severity, message, occurred_at, created_at
	FROM operational_events
	WHERE spacecraft_id = $1
	ORDER BY occurred_at DESC, id DESC
	`

	rows, err := r.database.Query(ctx, query, spacecraftID)

	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	defer rows.Close()

	events := make([]Event, 0)

	for rows.Next() {
		var event Event

		if err := rows.Scan(
			&event.ID,
			&event.SpacecraftID,
			&event.TelemetrySampleID,
			&event.Code,
			&event.Severity,
			&event.Message,
			&event.OccurredAt,
			&event.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return events, nil
}
