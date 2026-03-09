ALTER TABLE jobs ADD COLUMN file_path VARCHAR(500);
ALTER TABLE jobs DROP COLUMN source_id;
DROP TABLE sources;
DROP TYPE source_type;
