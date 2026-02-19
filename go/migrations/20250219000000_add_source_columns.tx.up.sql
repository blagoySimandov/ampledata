ALTER TABLE jobs
ADD COLUMN IF NOT EXISTS source_columns JSONB DEFAULT '[]'::jsonb;
