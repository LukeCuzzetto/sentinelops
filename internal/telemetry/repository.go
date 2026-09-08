package telemetry

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("telemetry sample not found")

type Repository struct {
	database *pgxpool.Pool
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
		return Sample{}, fmt.Errorf(
			"create telemetry sample: %w",
			err,
		)
	}

	return sample, nil
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
			"%w: spacecraft if %d",
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
