package telemetry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound              = errors.New("telemetry sample not found")
	ErrSpacecraftNotFound    = errors.New("spacecraft not found")
	ErrSequenceAlreadyExists = errors.New("telemetry sample with sequence number already exists")
	ErrSequenceConflict      = errors.New("telemetry sample with sequence number already exists with different data")
)

type Repository struct {
	database *pgxpool.Pool
}

type IngestResult struct {
	Sample  Sample
	Created bool
}

func NewRepository(
	database *pgxpool.Pool,
) *Repository {
	return &Repository{
		database: database,
	}
}

func (repository *Repository) CreateSample(
	ctx context.Context,
	params CreateSampleParams,
) (Sample, error) {
	const query = `
		INSERT INTO telemetry_samples (
			spacecraft_id,
			sequence_number,
			source_timestamp,
			battery_voltage,
			battery_soc_percent,
			temperature_c,
			mode
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7
		)
		RETURNING
			id,
			spacecraft_id,
			sequence_number,
			source_timestamp,
			received_at,
			battery_voltage,
			battery_soc_percent,
			temperature_c,
			mode
	`

	var sample Sample

	err := repository.database.QueryRow(
		ctx,
		query,
		params.SpacecraftID,
		params.SequenceNumber,
		params.SourceTimestamp,
		params.BatteryVoltage,
		params.BatterySOCPercent,
		params.TemperatureC,
		params.Mode,
	).Scan(
		&sample.ID,
		&sample.SpacecraftID,
		&sample.SequenceNumber,
		&sample.SourceTimestamp,
		&sample.ReceivedAt,
		&sample.BatteryVoltage,
		&sample.BatterySOCPercent,
		&sample.TemperatureC,
		&sample.Mode,
	)

	if err != nil {
		var postgresError *pgconn.PgError

		if errors.As(err, &postgresError) {
			switch postgresError.Code {
			case "23503":
				return Sample{}, fmt.Errorf(
					"%w: spacecraft id %d",
					ErrSpacecraftNotFound,
					params.SpacecraftID,
				)

			case "23505":
				return Sample{}, fmt.Errorf(
					"%w: spacecraft id %d, sequence number %d",
					ErrSequenceAlreadyExists,
					params.SpacecraftID,
					params.SequenceNumber,
				)
			}
		}

		return Sample{}, fmt.Errorf(
			"create telemetry sample: %w",
			err,
		)
	}

	return sample, nil
}

func (repository *Repository) IngestSample(
	ctx context.Context,
	params CreateSampleParams,
) (IngestResult, error) {
	created, err := repository.CreateSample(ctx, params)
	if err == nil {
		return IngestResult{
			Sample:  created,
			Created: true,
		}, nil
	}

	if !errors.Is(err, ErrSequenceAlreadyExists) {
		return IngestResult{}, err
	}

	existing, err := repository.getSampleBySequenceNumber(
		ctx,
		params.SpacecraftID,
		params.SequenceNumber,
	)
	if err != nil {
		return IngestResult{}, fmt.Errorf(
			"get existing telemetry sample after sequence conflict: %w",
			err,
		)
	}

	if sameSamplePayload(existing, params) {
		return IngestResult{
			Sample:  existing,
			Created: false,
		}, nil
	}

	return IngestResult{}, fmt.Errorf(
		"%w: spacecraft id %d, sequence number %d",
		ErrSequenceConflict,
		params.SpacecraftID,
		params.SequenceNumber,
	)
}

func (repository *Repository) ListSamplesBySpacecraftID(
	ctx context.Context,
	spacecraftID int64,
) ([]Sample, error) {
	const query = `
		SELECT
			id,
			spacecraft_id,
			sequence_number,
			source_timestamp,
			received_at,
			battery_voltage,
			battery_soc_percent,
			temperature_c,
			mode
		FROM telemetry_samples
		WHERE spacecraft_id = $1
		ORDER BY source_timestamp DESC, sequence_number DESC
	`

	rows, err := repository.database.Query(
		ctx,
		query,
		spacecraftID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"list telemetry samples by spacecraft ID: %w",
			err,
		)
	}

	defer rows.Close()

	samples := make([]Sample, 0)

	for rows.Next() {
		var sample Sample

		err := rows.Scan(
			&sample.ID,
			&sample.SpacecraftID,
			&sample.SequenceNumber,
			&sample.SourceTimestamp,
			&sample.ReceivedAt,
			&sample.BatteryVoltage,
			&sample.BatterySOCPercent,
			&sample.TemperatureC,
			&sample.Mode,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"scan telemetry sample: %w",
				err,
			)
		}

		samples = append(samples, sample)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate telemetry samples: %w",
			err,
		)
	}

	return samples, nil
}

func (repository *Repository) GetLatestSampleBySpacecraftID(
	ctx context.Context,
	spacecraftID int64,
) (Sample, error) {
	const query = `
		SELECT
			id,
			spacecraft_id,
			sequence_number,
			source_timestamp,
			received_at,
			battery_voltage,
			battery_soc_percent,
			temperature_c,
			mode
		FROM telemetry_samples
		WHERE spacecraft_id = $1
		ORDER BY source_timestamp DESC, sequence_number DESC
		LIMIT 1
	`

	var sample Sample

	err := repository.database.QueryRow(
		ctx,
		query,
		spacecraftID,
	).Scan(
		&sample.ID,
		&sample.SpacecraftID,
		&sample.SequenceNumber,
		&sample.SourceTimestamp,
		&sample.ReceivedAt,
		&sample.BatteryVoltage,
		&sample.BatterySOCPercent,
		&sample.TemperatureC,
		&sample.Mode,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Sample{}, fmt.Errorf(
			"%w: spacecraft id %d",
			ErrNotFound,
			spacecraftID,
		)
	}

	if err != nil {
		return Sample{}, fmt.Errorf(
			"get latest telemetry sample by spacecraft ID: %w",
			err,
		)
	}

	return sample, nil
}

func (repository *Repository) getSampleBySequenceNumber(
	ctx context.Context,
	spacecraftID int64,
	sequenceNumber int64,
) (Sample, error) {
	const query = `
		SELECT
			id,
			spacecraft_id,
			sequence_number,
			source_timestamp,
			received_at,
			battery_voltage,
			battery_soc_percent,
			temperature_c,
			mode
		FROM telemetry_samples
		WHERE spacecraft_id = $1 AND sequence_number = $2
	`

	var sample Sample

	err := repository.database.QueryRow(
		ctx,
		query,
		spacecraftID,
		sequenceNumber,
	).Scan(
		&sample.ID,
		&sample.SpacecraftID,
		&sample.SequenceNumber,
		&sample.SourceTimestamp,
		&sample.ReceivedAt,
		&sample.BatteryVoltage,
		&sample.BatterySOCPercent,
		&sample.TemperatureC,
		&sample.Mode,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Sample{}, fmt.Errorf(
			"%w: spacecraft id %d, sequence number %d",
			ErrNotFound,
			spacecraftID,
			sequenceNumber,
		)
	}

	if err != nil {
		return Sample{}, fmt.Errorf(
			"get telemetry sample by spacecraft ID and sequence number: %w",
			err,
		)
	}

	return sample, nil
}

func sameSamplePayload(existing Sample, params CreateSampleParams) bool {
	sourceTimestamp := params.SourceTimestamp.UTC().Truncate(time.Microsecond)

	return existing.SourceTimestamp.Equal(sourceTimestamp) &&
		existing.BatteryVoltage == params.BatteryVoltage &&
		existing.BatterySOCPercent == params.BatterySOCPercent &&
		existing.TemperatureC == params.TemperatureC &&
		existing.Mode == params.Mode
}
