CREATE TYPE source_type AS ENUM ('csv_upload');

CREATE TABLE sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR NOT NULL REFERENCES users(id),
    type source_type NOT NULL,
    metadata JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sources_user_id ON sources (user_id);

ALTER TABLE jobs ADD COLUMN source_id UUID REFERENCES sources(id);
ALTER TABLE jobs DROP COLUMN file_path;
