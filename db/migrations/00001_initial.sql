-- +goose Up
-- +goose StatementBegin
CREATE TABLE incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    slack_channel_id VARCHAR NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('open', 'resolved', 'cancelled')),
    severity VARCHAR(10) NOT NULL CHECK (severity IN ('sev1', 'sev2', 'sev3')),
    title TEXT NOT NULL,
    description TEXT,
    started_by VARCHAR NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    resolved_by VARCHAR,
    resolved_at TIMESTAMP WITH TIME ZONE,
    export_url TEXT,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE timeline_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    incident_id UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    event_type VARCHAR(50) NOT NULL CHECK (
        event_type IN (
            'incident_started',
            'severity_change',
            'message_reaction',
            'file_upload',
            'resolved',
            'custom'
        )
    ),
    slack_user_id VARCHAR NOT NULL,
    slack_message_ts VARCHAR,
    metadata JSONB DEFAULT '{}'::jsonb
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE timeline_events;
DROP TABLE incidents;
-- +goose StatementEnd
