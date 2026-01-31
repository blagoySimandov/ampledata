-- Add back the key_column column
ALTER TABLE jobs
ADD COLUMN IF NOT EXISTS key_column VARCHAR(255);

--bun:split

-- Migrate data back from key_columns to key_column (take first element)
UPDATE jobs
SET key_column = CASE
    WHEN key_columns IS NOT NULL AND jsonb_array_length(key_columns) > 0 THEN key_columns->>0
    ELSE NULL
END
WHERE key_column IS NULL;

--bun:split

-- Drop the key_columns column
ALTER TABLE jobs
DROP COLUMN IF EXISTS key_columns;
