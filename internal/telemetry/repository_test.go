package telemetry

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/LukeCuzzetto/sentinelops/internal/spacecraft"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRepositoryCreateSample(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL environment variable is not set")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(
		ctx, databaseURL,
	)
	if err != nil {
		t.Fatalf("create database pool: %v", err)
	}

	defer pool.Close()

	spacecraftRepository := spacecraft.NewRepository(pool)

	spacecraftName := "TELEMETRY-TEST-" + time.Now().Format("20060102150405.000000000")

	createdSpacecraft, err := spacecraftRepository.CreateSpacecraft(ctx, spacecraftName)

	if err != nil {
		t.Fatalf("create test spacecraft: %v", err)
	}
	defer func() {
		_, _ = pool.Exec(
			ctx, "DELETE FROM spacecraft WHERE id = $1",
			createdSpacecraft.ID,
		)
	}()

	repository := NewRepository(pool)

	sourceTimestamp := time.Now().UTC().Truncate(time.Microsecond)

	params := CreateSampleParams{
		SpacecraftID:      createdSpacecraft.ID,
		SequenceNumber:    101,
		SourceTimestamp:   sourceTimestamp,
		BatteryVoltage:    8.2,
		BatterySOCPercent: 76.4,
		TemperatureC:      21.7,
		Mode:              "nominal",
	}

	created, err := repository.CreateSample(ctx, params)

	if err != nil {
		t.Fatalf("create telemetry sample: %v", err)
	}

	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM telemetry_samples WHERE id = $1", created.ID)
	}()

	if created.ID == 0 {
		t.Error("expected generate telemetry ID")

	}

	if created.SpacecraftID != createdSpacecraft.ID {
		t.Errorf("expected spacecraft ID %d, got %d", createdSpacecraft.ID, created.SpacecraftID)
	}

	if created.SequenceNumber != params.SequenceNumber {
		t.Errorf("expected sequence number %d, got %d", params.SequenceNumber, created.SequenceNumber)
	}

	if !created.SourceTimestamp.Equal(params.SourceTimestamp) {
		t.Errorf("expected source timestamp %v, got %v", params.SourceTimestamp, created.SourceTimestamp)

	}

	if created.BatteryVoltage != params.BatteryVoltage {
		t.Errorf("expected battery voltage %.2f, got %.2f", params.BatteryVoltage, created.BatteryVoltage)

	}

	if created.BatterySOCPercent != params.BatterySOCPercent {
		t.Errorf("expected battery SOC %.2f, got %.2f", params.BatterySOCPercent, created.BatterySOCPercent)

	}

	if created.TemperatureC != params.TemperatureC {
		t.Errorf("expected temperature %.2f, got %.2f", params.TemperatureC, created.TemperatureC)
	}

	if created.Mode != params.Mode {
		t.Errorf("expected mode %q, got %q", params.Mode, created.Mode)
	}

}
