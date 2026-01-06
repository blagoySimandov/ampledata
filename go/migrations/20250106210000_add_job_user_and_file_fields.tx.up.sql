ALTER TABLE jobs
ADD COLUMN IF NOT EXISTS user_id VARCHAR(255) NOT NULL DEFAULT '';

--bun:split

ALTER TABLE jobs
ADD COLUMN IF NOT EXISTS file_path VARCHAR(500) DEFAULT '';

--bun:split

ALTER TABLE jobs
ADD COLUMN IF NOT EXISTS key_column VARCHAR(255);

--bun:split

ALTER TABLE jobs
ADD COLUMN IF NOT EXISTS columns_metadata JSONB;

--bun:split

ALTER TABLE jobs
ADD COLUMN IF NOT EXISTS entity_type VARCHAR(255);

--bun:split

ALTER TABLE jobs ALTER COLUMN started_at DROP NOT NULL;

--bun:split

ALTER TABLE jobs ALTER COLUMN started_at DROP DEFAULT;

--bun:split

ALTER TABLE jobs ALTER COLUMN status SET DEFAULT 'PENDING';

--bun:split

CREATE INDEX IF NOT EXISTS idx_jobs_user_id ON jobs(user_id);

--bun:split

CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
