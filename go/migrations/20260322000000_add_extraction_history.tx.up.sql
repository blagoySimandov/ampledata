ALTER TABLE row_states
ADD COLUMN IF NOT EXISTS extraction_history JSONB;
