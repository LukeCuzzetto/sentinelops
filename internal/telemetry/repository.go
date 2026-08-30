package telemetry

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

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
