-- Add new key_columns JSONB column
ALTER TABLE jobs
ADD COLUMN IF NOT EXISTS key_columns JSONB;

--bun:split

-- Migrate existing data from key_column to key_columns
UPDATE jobs
SET key_columns = CASE
    WHEN key_column IS NOT NULL AND key_column != '' THEN jsonb_build_array(key_column)
    ELSE NULL
END
WHERE key_columns IS NULL;

--bun:split

-- Drop the old key_column column
ALTER TABLE jobs
DROP COLUMN IF EXISTS key_column;
