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

func TestRepositoryCreateSampleInvalidSpacecraftID(t *testing.T) {
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

	spacecraftName := "HISTORY-TEST-" + time.Now().Format("20060102150405.000000000")

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

	baseTime := time.Now().UTC().Truncate(time.Microsecond)

	first, err := repository.CreateSample(ctx, CreateSampleParams{
		SpacecraftID:      createdSpacecraft.ID,
		SequenceNumber:    101,
		SourceTimestamp:   baseTime,
		BatteryVoltage:    8.2,
		BatterySOCPercent: 80,
		TemperatureC:      20,
		Mode:              "nominal",
	})

	if err != nil {
		t.Fatalf("create first telemetry sample: %v", err)
	}

	third, err := repository.CreateSample(ctx, CreateSampleParams{
		SpacecraftID:      createdSpacecraft.ID,
		SequenceNumber:    103,
		SourceTimestamp:   baseTime.Add(2 * time.Minute),
		BatteryVoltage:    8.0,
		BatterySOCPercent: 78,
		TemperatureC:      22,
		Mode:              "nominal",
	})

	if err != nil {
		t.Fatalf("create third telemetry sample: %v", err)
	}

	second, err := repository.CreateSample(ctx, CreateSampleParams{
		SpacecraftID:      createdSpacecraft.ID,
		SequenceNumber:    102,
		SourceTimestamp:   baseTime.Add(time.Minute),
		BatteryVoltage:    8.1,
		BatterySOCPercent: 79,
		TemperatureC:      21,
		Mode:              "nominal",
	})

	if err != nil {
		t.Fatalf("create second telemetry sample: %v", err)
	}

	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM telemetry_samples WHERE id = $1", first.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM telemetry_samples WHERE id = $1", second.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM telemetry_samples WHERE id = $1", third.ID)
	}()

	samples, err := repository.ListSamplesBySpacecraftID(
		ctx, createdSpacecraft.ID,
	)

	if err != nil {
		t.Fatalf("list telemetry samples by spacecraft ID: %v", err)
	}

	if len(samples) != 3 {
		t.Fatalf("expected 3 telemetry samples, got %d", len(samples))
	}

	expectedSequenceNumbers := []int64{103, 102, 101}

	for index, expected := range expectedSequenceNumbers {
		if samples[index].SequenceNumber != expected {
			t.Errorf("expected sequence number %d, got %d", expected, samples[index].SequenceNumber)
		}
	}
}

func TestRepositoryGetLatestSampleBySpacecraftID(t *testing.T) {
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

	spacecraftName := "LATEST-TEST-" + time.Now().Format("20060102150405.000000000")

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

	baseTime := time.Now().UTC().Truncate(time.Microsecond)

	older, err := repository.CreateSample(ctx, CreateSampleParams{
		SpacecraftID:      createdSpacecraft.ID,
		SequenceNumber:    102,
		SourceTimestamp:   baseTime.Add(time.Minute),
		BatteryVoltage:    8.1,
		BatterySOCPercent: 75,
		TemperatureC:      21,
		Mode:              "nominal",
	},
	)

	if err != nil {
		t.Fatalf("create older telemetry sample: %v", err)
	}

	newer, err := repository.CreateSample(ctx, CreateSampleParams{
		SpacecraftID:      createdSpacecraft.ID,
		SequenceNumber:    103,
		SourceTimestamp:   baseTime.Add(2 * time.Minute),
		BatteryVoltage:    8.0,
		BatterySOCPercent: 74,
		TemperatureC:      22,
		Mode:              "nominal",
	},
	)

	if err != nil {
		t.Fatalf("create newer telemetry sample: %v", err)
	}

	late, err := repository.CreateSample(ctx, CreateSampleParams{
		SpacecraftID:      createdSpacecraft.ID,
		SequenceNumber:    101,
		SourceTimestamp:   baseTime,
		BatteryVoltage:    8.2,
		BatterySOCPercent: 76,
		TemperatureC:      20,
		Mode:              "nominal",
	},
	)

	if err != nil {
		t.Fatalf("create late telemetry sample: %v", err)
	}

	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM telemetry_samples WHERE id = $1", older.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM telemetry_samples WHERE id = $1", newer.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM telemetry_samples WHERE id = $1", late.ID)
	}()

	latest, err := repository.GetLatestSampleBySpacecraftID(
		ctx, createdSpacecraft.ID,
	)

	if err != nil {
		t.Fatalf("get latest telemetry sample by spacecraft ID: %v", err)
	}

	if latest.SequenceNumber != 103 {
		t.Errorf("expected latest sample sequence number 103, got %d", latest.SequenceNumber)
	}
}
