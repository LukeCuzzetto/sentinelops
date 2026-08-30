-- +goose up

CREATE TABLE telemetry_samples (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    spacecraft_id BIGINT NOT NULL
        REFERENCES spacecraft(id),

    sequence_number BIGINT NOT NULL
        CHECK (sequence_number >= 0),

        source_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
        received_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

        battery_voltage DOUBLE PRECISION NOT NULL
            CHECK (battery_voltage >= 0),

        battery_soc_percent DOUBLE PRECISION NOT NULL
            CHECK (
                battery_soc_percent >= 0
                AND battery_soc_percent <= 100
            ),

        temperature_c DOUBLE PRECISION NOT NULL,

        mode TEXT NOT NULL,

        UNIQUE (spacecraft_id, sequence_number)
);

CREATE INDEX telemetry_samples_spacecraft_source_timestamp_idx
    ON telemetry_samples (
        spacecraft_id,
        source_timestamp DESC
    );

-- +goose down

DROP TABLE telemetry_samples;