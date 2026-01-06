DROP INDEX IF EXISTS idx_jobs_status;

--bun:split

DROP INDEX IF EXISTS idx_jobs_user_id;

--bun:split

ALTER TABLE jobs ALTER COLUMN status SET DEFAULT 'RUNNING';

--bun:split

ALTER TABLE jobs ALTER COLUMN started_at SET DEFAULT current_timestamp;

--bun:split

ALTER TABLE jobs ALTER COLUMN started_at SET NOT NULL;

--bun:split

ALTER TABLE jobs DROP COLUMN IF EXISTS entity_type;

--bun:split

ALTER TABLE jobs DROP COLUMN IF EXISTS columns_metadata;

--bun:split

ALTER TABLE jobs DROP COLUMN IF EXISTS key_column;

--bun:split

ALTER TABLE jobs DROP COLUMN IF EXISTS file_path;

--bun:split

ALTER TABLE jobs DROP COLUMN IF EXISTS user_id;
