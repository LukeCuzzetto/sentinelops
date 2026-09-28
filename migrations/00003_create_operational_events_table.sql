-- +goose up

CREATE TABLE operational_events (

    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    spacecraft_id BIGINT NOT NULL
        REFERENCES spacecraft(id),


    telemetry_sample_id BIGINT
        REFERENCES TELEMETRY_SAMPLES(id),
    
    code TEXT NOT NULL,

    severity TEXT NOT NULL
        CHECK (
            severity IN ('info', 'warning', 'critical')
        ),

    message TEXT NOT NULL,

    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX operational_eventts_spacecraft_occurred_at_idx ON operational_events(spacecraft_id, occurred_at DESC);

-- +goose down

DROP TABLE operational_events;