package events

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/LukeCuzzetto/sentinelops/internal/spacecraft"
	"github.com/LukeCuzzetto/sentinelops/internal/telemetry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRepositoryCreateEvent(t *testing.T) {

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL environment variable is not set")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, databaseURL)

	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	spaceRepository := spacecraft.NewRepository(pool)

	spacecraftName := "EVENT-TEST-" + time.Now().Format("20060102150405")

	createdSpacecraft, err := spaceRepository.CreateSpacecraft(ctx, spacecraftName)

	if err != nil {
		t.Fatalf("create spacecraft: %v", err)
	}

	telemetryRepository := telemetry.NewRepository(pool)

	sourceTimestamp := time.Now().UTC().Truncate(time.Microsecond)

	createdSample, err := telemetryRepository.CreateSample(ctx, telemetry.CreateSampleParams{

		SpacecraftID:      createdSpacecraft.ID,
		SequenceNumber:    1,
		SourceTimestamp:   sourceTimestamp,
		BatteryVoltage:    8.2,
		BatterySOCPercent: 10.0,
		TemperatureC:      21,
		Mode:              "nominal",
	})
	if err != nil {
		t.Fatalf("failed to create telemetry sample: %v", err)
	}

	repository := NewRepository(pool)

	telemetrySampleID := createdSample.ID

	eventParams := CreateEventParams{
		SpacecraftID:      createdSpacecraft.ID,
		TelemetrySampleID: &telemetrySampleID,
		Code:              "batter_soc_critical",
		Severity:          SeverityCritical,
		Message:           "battery state of charge is critically low",
		OccurredAt:        sourceTimestamp,
	}

	createdEvent, err := repository.CreateEvent(ctx, eventParams)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM operational_events WHERE id = $1", createdEvent.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM telemetry_samples WHERE id = $1", createdSample.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM spacecraft WHERE id = $1", createdSpacecraft.ID)

	}()

	if createdEvent.ID == 0 {
		t.Error("expected generated event ID")

	}

	if createdEvent.SpacecraftID != createdSpacecraft.ID {
		t.Errorf("expected spacecraft ID %d, got %d", createdSpacecraft.ID, createdEvent.SpacecraftID)
	}

	if createdEvent.TelemetrySampleID == nil {
		t.Fatal("expected telemetry sample ID")
	}
}

func TestRepositoryListEventsBySpacecraftID(t *testing.T) {

	// ARRANGE the world we need for the test, spacecraft and the 3 events to be created
	// Always DatabaseURL from the environment, then context, then connect to the database,
	// then create a spacecraft, then create 3 events for that spacecraft

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL environment variable is not set")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, databaseURL)

	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	spaceRepository := spacecraft.NewRepository(pool)

	spacecraftName := "EVENT-TEST-" + time.Now().Format("20060102150405")

	createdSpacecraft, err := spaceRepository.CreateSpacecraft(ctx, spacecraftName)

	if err != nil {
		t.Fatalf("create spacecraft: %v", err)
	}

	defer func() {
		_, _ = pool.Exec(
			ctx,
			"DELETE FROM spacecraft WHERE id = $1",
			createdSpacecraft.ID,
		)
	}()

	repository := NewRepository(pool)

	baseTime := time.Now().
		UTC().
		Truncate(time.Microsecond)

	first, err := repository.CreateEvent(
		ctx,
		CreateEventParams{
			SpacecraftID: createdSpacecraft.ID,
			Code:         "event_first",
			Severity:     SeverityWarning,
			Message:      "first event",
			OccurredAt:   baseTime,
		},
	)
	if err != nil {
		t.Fatalf(
			"create first event: %v",
			err,
		)
	}

	third, err := repository.CreateEvent(
		ctx,
		CreateEventParams{
			SpacecraftID: createdSpacecraft.ID,
			Code:         "event_third",
			Severity:     SeverityCritical,
			Message:      "third event",
			OccurredAt:   baseTime.Add(2 * time.Minute),
		},
	)
	if err != nil {
		t.Fatalf(
			"create third event: %v",
			err,
		)
	}

	second, err := repository.CreateEvent(
		ctx,
		CreateEventParams{
			SpacecraftID: createdSpacecraft.ID,
			Code:         "event_second",
			Severity:     SeverityWarning,
			Message:      "second event",
			OccurredAt:   baseTime.Add(time.Minute),
		},
	)
	if err != nil {
		t.Fatalf(
			"create second event: %v",
			err,
		)
	}

	defer func() {
		_, _ = pool.Exec(
			ctx,
			"DELETE FROM operational_events WHERE id IN ($1, $2, $3)",
			first.ID,
			second.ID,
			third.ID,
		)
	}()

	// ACT - list the events for the spacecraft

	events, err := repository.ListEventsBySpacecraftID(ctx, createdSpacecraft.ID)
	if err != nil {
		t.Fatalf("failed to list events: %v", err)
	}

	// ASSERT - verify the events are returned in the expected order

	if len(events) != 3 {
		t.Fatalf(
			"expected 3 events, got %d",
			len(events),
		)
	}

	expectedCodes := []string{
		"event_third",
		"event_second",
		"event_first",
	}

	for index, expected := range expectedCodes {
		if events[index].Code != expected {
			t.Errorf(
				"expected code %q at index %d, got %q",
				expected,
				index,
				events[index].Code,
			)
		}
	}
}
